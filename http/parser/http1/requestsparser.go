package http1

import (
	"indigo/errors"
	"indigo/http/headers"
	methods "indigo/http/method"
	"indigo/http/parser"
	"indigo/http/proto"
	"indigo/http/url"
	"indigo/internal"
	"indigo/settings"
	"indigo/types"
)

type httpRequestsParser struct {
	state   parserState
	request *types.Request

	settings settings.Settings

	contentLength           uint
	lengthCountdown         uint
	closeConnection         bool
	chunkedTransferEncoding bool
	chunkedBodyParser       chunkedBodyParser

	startLineBuff  []byte
	offset         int
	urlEncodedChar uint8

	headerBuff     []byte
	addHeaderValue headers.ValueAppender

	body *internal.BodyGateway
}

func NewHTTPRequestsParser(
	request *types.Request, body *internal.BodyGateway,
	startLineBuff, headerBuff []byte, settings settings.Settings,
) parser.HTTPRequestsParser {
	return &httpRequestsParser{
		state:   eMethod,
		request: request,

		settings: settings,

		startLineBuff: startLineBuff,
		headerBuff:    headerBuff,

		body: body,
	}
}

func (p *httpRequestsParser) Parse(data []byte) (state parser.RequestState, extra []byte, err error) {
	if len(data) == 0 {
		p.body.WriteErr(errors.ErrCloseConnection)

		return parser.ConnectionClose, nil, nil
	}

	if p.state == eBody {
		var done bool
		done, extra, err = p.parseBody(data)
		if err != nil {
			p.body.WriteErr(err)

			return parser.Error, nil, err
		} else if done {
			p.body.Data <- nil
			p.reset()

			return parser.BodyCompleted, extra, err
		}

		return parser.Pending, extra, nil
	}

	for i := range data {
		switch p.state {
		case eMethod:
			if data[i] == ' ' {
				p.request.Method = methods.Parse(internal.B2S(p.startLineBuff))

				if p.request.Method == methods.Unknown {
					return parser.Error, nil, errors.ErrBadRequest
				}

				p.offset = len(p.startLineBuff)
				p.state = ePath
				continue
			}

			if len(p.startLineBuff) > len("CONNECT") { // the longest method, trust me
				return parser.Error, nil, errors.ErrBadRequest
			}

			p.startLineBuff = append(p.startLineBuff, data[i])
		case ePath:
			switch data[i] {
			case ' ':
				p.request.Path = url.Path(internal.B2S(p.startLineBuff[p.offset:]))
				p.offset = len(p.startLineBuff)
				p.state = eProto
			case '%':
				p.state = ePathDecode1Char
			case '?':
				p.offset = len(p.startLineBuff)
				p.state = eQuery
			case '#':
				p.offset = len(p.startLineBuff)
				p.state = eFragment
			default:
				if len(p.startLineBuff) >= int(p.settings.URLBuffSize.Maximal) {
					return parser.Error, nil, errors.ErrURLTooLong
				}

				p.startLineBuff = append(p.startLineBuff, data[i])
			}
		case ePathDecode1Char:
			if !isHex(data[i]) {
				return parser.Error, nil, errors.ErrURLDecoding
			}

			p.urlEncodedChar = unHex(data[i]) << 4
			p.state = ePathDecode2Char
		case ePathDecode2Char:
			if !isHex(data[i]) {
				return parser.Error, nil, errors.ErrURLDecoding
			}
			if len(p.startLineBuff) >= int(p.settings.URLBuffSize.Maximal) {
				return parser.Error, nil, errors.ErrURLTooLong
			}

			p.startLineBuff = append(p.startLineBuff, p.urlEncodedChar|unHex(data[i]))
			p.urlEncodedChar = 0
			p.state = ePath
		case eQuery:
			switch data[i] {
			case ' ':
				p.request.Query = url.NewQuery(p.startLineBuff[p.offset:])
				p.offset = len(p.startLineBuff)
				p.state = eProto
			case '#':
				p.offset = len(p.startLineBuff)
				p.state = eFragment
			case '%':
				p.state = eQueryDecode1Char
			case '+':
				if len(p.startLineBuff) >= int(p.settings.URLBuffSize.Maximal) {
					return parser.Error, nil, errors.ErrURLTooLong
				}

				p.startLineBuff = append(p.startLineBuff, ' ')
			default:
				if len(p.startLineBuff) >= int(p.settings.URLBuffSize.Maximal) {
					return parser.Error, nil, errors.ErrURLTooLong
				}

				p.startLineBuff = append(p.startLineBuff, data[i])
			}
		case eQueryDecode1Char:
			if !isHex(data[i]) {
				return parser.Error, nil, errors.ErrURLDecoding
			}

			p.urlEncodedChar = unHex(data[i]) << 4
			p.state = eQueryDecode2Char
		case eQueryDecode2Char:
			if !isHex(data[i]) {
				return parser.Error, nil, errors.ErrURLDecoding
			}
			if len(p.startLineBuff) >= int(p.settings.URLBuffSize.Maximal) {
				return parser.Error, nil, errors.ErrURLTooLong
			}

			p.startLineBuff = append(p.startLineBuff, p.urlEncodedChar|unHex(data[i]))
			p.urlEncodedChar = 0
			p.state = eQuery
		case eFragment:
			switch data[i] {
			case ' ':
				p.request.Fragment = url.Fragment(internal.B2S(p.startLineBuff[p.offset:]))
				p.offset = len(p.startLineBuff)
				p.state = eProto
			case '%':
				p.state = eFragmentDecode1Char
			default:
				if len(p.startLineBuff) >= int(p.settings.URLBuffSize.Maximal) {
					return parser.Error, nil, errors.ErrURLTooLong
				}

				p.startLineBuff = append(p.startLineBuff, data[i])
			}
		case eFragmentDecode1Char:
			if !isHex(data[i]) {
				return parser.Error, nil, errors.ErrURLDecoding
			}

			p.urlEncodedChar = unHex(data[i]) << 4
			p.state = eFragmentDecode2Char
		case eFragmentDecode2Char:
			if !isHex(data[i]) {
				return parser.Error, nil, errors.ErrURLDecoding
			}
			if len(p.startLineBuff) >= int(p.settings.URLBuffSize.Maximal) {
				return parser.Error, nil, errors.ErrURLTooLong
			}

			p.startLineBuff = append(p.startLineBuff, p.urlEncodedChar|unHex(data[i]))
			p.urlEncodedChar = 0
			p.state = eFragment
		case eProto:
			switch data[i] {
			case '\r':
				p.request.Proto = proto.Parse(internal.B2S(p.startLineBuff[p.offset:]))
				if p.request.Proto == proto.Unknown {
					return parser.Error, nil, errors.ErrUnsupportedProtocol
				}
				p.state = eProtoCR
			case '\n':
				p.request.Proto = proto.Parse(internal.B2S(p.startLineBuff[p.offset:]))
				if p.request.Proto == proto.Unknown {
					return parser.Error, nil, errors.ErrUnsupportedProtocol
				}
				p.state = eProtoCRLF
			default:
				p.startLineBuff = append(p.startLineBuff, data[i])
			}
		case eProtoCR:
			if data[i] != '\n' {
				return parser.Error, nil, errors.ErrBadRequest
			}

			p.state = eProtoCRLF
		case eProtoCRLF:
			switch data[i] {
			case '\r':
				p.state = eProtoCRLFCR
			case '\n':
				p.reset()

				return parser.RequestCompleted, data[i+1:], nil
			default:
				// headers are here. I have to have a buffer for header key, and after receiving it,
				// get an appender from headers manager (and keep it in httpRequestsParser struct)
				p.state = eHeaderKey
				p.headerBuff = append(p.headerBuff, data[i]|0x20)
			}
		case eProtoCRLFCR:
			switch data[i] {
			case '\n':
				// no request body because even no headers
				p.reset()

				return parser.RequestCompleted, data[i+1:], nil
			default:
				return parser.Error, nil, errors.ErrBadRequest
			}
		case eHeaderKey:
			switch data[i] {
			case ':':
				p.addHeaderValue, err = p.request.Headers.Set(p.headerBuff)
				if err != nil {
					return parser.Error, nil, err
				}
				p.state = eHeaderColon
			case '\r', '\n':
				return parser.Error, nil, errors.ErrBadRequest
			default:
				if len(p.headerBuff) >= int(p.settings.HeaderKeyBuffSize.Maximal) {
					return parser.Error, nil, errors.ErrRequestEntityTooLarge
				}

				p.headerBuff = append(p.headerBuff, data[i]|0x20)
			}
		case eHeaderColon:
			switch data[i] {
			case '\r', '\n':
				return parser.Error, nil, errors.ErrBadRequest
			case ' ':
			default:
				p.addHeaderValue(data[i])
			}

			p.state = eHeaderValue
		case eHeaderValue:
			switch data[i] {
			case '\r':
				p.state = eHeaderValueCR
			case '\n':
				p.state = eHeaderValueCRLF
			default:
				p.addHeaderValue(data[i])
			}
		case eHeaderValueCR:
			switch data[i] {
			case '\n':
				p.state = eHeaderValueCRLF
			default:
				return parser.Error, nil, errors.ErrBadRequest
			}
		case eHeaderValueCRLF:
			switch internal.B2S(p.headerBuff) {
			case "content-length":
				header, _ := p.request.Headers.Get("content-length")
				p.contentLength, err = parseUint(header.Bytes())
				if err != nil {
					return parser.Error, nil, err
				}
				if p.contentLength > uint(p.settings.BodyLength.Maximal) {

				}
				p.lengthCountdown = p.contentLength
			case "connection":
				header, _ := p.request.Headers.Get("connection")
				p.closeConnection = header.String() == "close"
			case "transfer-encoding":
				header, _ := p.request.Headers.Get("transfer-encoding")
				// TODO: parse header value
				p.chunkedTransferEncoding = header.String() == "chunked"
			}

			p.headerBuff = p.headerBuff[:0]

			switch data[i] {
			case '\n':
				if p.contentLength == 0 && !p.chunkedTransferEncoding {
					p.reset()

					return parser.RequestCompleted, data[i+1:], nil
				}

				p.state = eBody

				return parser.HeadersCompleted, data[i+1:], nil
			case '\r':
				p.state = eHeaderValueCRLFCR
			default:
				p.headerBuff = append(p.headerBuff[:0], data[i]|0x20)
				p.state = eHeaderKey
			}
		case eHeaderValueCRLFCR:
			switch data[i] {
			case '\n':
				if p.contentLength == 0 && !p.chunkedTransferEncoding {
					p.reset()

					return parser.RequestCompleted, data[i+1:], nil
				}

				p.state = eBody

				return parser.HeadersCompleted, data[i+1:], nil
			default:
				return parser.Error, nil, errors.ErrBadRequest
			}
		}
	}

	return parser.Pending, nil, nil
}

func (p *httpRequestsParser) parseBody(b []byte) (done bool, extra []byte, err error) {
	if p.chunkedTransferEncoding {
		panic("Chunked transfer encoding is not implemented!")
	}

	if p.lengthCountdown <= uint(len(b)) {
		p.body.Data <- b[:p.lengthCountdown]
		<-p.body.Data
		p.lengthCountdown = 0

		return true, b[p.lengthCountdown:], p.body.Err
	}

	p.body.Data <- b
	<-p.body.Data
	p.lengthCountdown -= uint(len(b))

	return false, nil, p.body.Err
}

func (p *httpRequestsParser) FinalizeBody() {
	p.body.Data <- nil
}

func (p *httpRequestsParser) reset() {
	p.startLineBuff = p.startLineBuff[:0]
	p.offset = 0
	p.headerBuff = p.headerBuff[:0]
	p.contentLength = 0
	p.chunkedTransferEncoding = false
	p.state = eMethod
}
