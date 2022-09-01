package http1

import (
	"indigo/errors"
	"indigo/http/headers"
	methods "indigo/http/method"
	"indigo/http/parser"
	"indigo/http/proto"
	"indigo/internal"
	"indigo/settings"
	"indigo/types"
)

// httpRequestsParser is a stream-based http requests parser. It modifies
// request object by pointer in performance purposes. Decodes url-encoded
// values by its own, you can see that by presented states ePathDecode1Char,
// ePathDecode2Char, etc. When headers are parsed, parser returns state
// parser.HeadersCompleted to notify http server about this, attaching all
// the pending data as an extra. Body must be processed separately
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
	headersManager *headers.Manager

	body *internal.BodyGateway
}

func NewHTTPRequestsParser(
	request *types.Request, body *internal.BodyGateway,
	startLineBuff, headerBuff []byte, settings settings.Settings,
	manager *headers.Manager,
) parser.HTTPRequestsParser {
	return &httpRequestsParser{
		state:   eMethod,
		request: request,

		chunkedBodyParser: newChunkedBodyParser(body, settings),
		settings:          settings,
		startLineBuff:     startLineBuff,
		headerBuff:        headerBuff,
		headersManager:    manager,

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
				if len(p.startLineBuff) == p.offset {
					return parser.Error, nil, errors.ErrBadRequest
				}

				p.request.Path = internal.B2S(p.startLineBuff[p.offset:])
				p.offset = len(p.startLineBuff)
				p.state = eProto
			case '%':
				p.state = ePathDecode1Char
			case '?':
				p.request.Path = internal.B2S(p.startLineBuff[p.offset:])
				if len(p.request.Path) == 0 {
					p.request.Path = "/"
				}

				p.offset = len(p.startLineBuff)
				p.state = eQuery
			case '#':
				p.request.Path = internal.B2S(p.startLineBuff[p.offset:])
				if len(p.request.Path) == 0 {
					p.request.Path = "/"
				}

				p.offset = len(p.startLineBuff)
				p.state = eFragment
			case '\x00', '\n', '\r', '\t', '\b', '\a', '\v', '\f':
				// request path MUST NOT include any non-printable characters
				return parser.Error, nil, errors.ErrBadRequest
			default:
				if uint16(len(p.startLineBuff)) >= p.settings.URL.Length.Maximal {
					return parser.Error, nil, errors.ErrURITooLong
				}

				p.startLineBuff = append(p.startLineBuff, data[i])
			}
		case ePathDecode1Char:
			if !isHex(data[i]) {
				return parser.Error, nil, errors.ErrURIDecoding
			}

			p.urlEncodedChar = unHex(data[i]) << 4
			p.state = ePathDecode2Char
		case ePathDecode2Char:
			if !isHex(data[i]) {
				return parser.Error, nil, errors.ErrURIDecoding
			}
			if uint16(len(p.startLineBuff)) >= p.settings.URL.Length.Maximal {
				return parser.Error, nil, errors.ErrURITooLong
			}

			p.startLineBuff = append(p.startLineBuff, p.urlEncodedChar|unHex(data[i]))
			p.urlEncodedChar = 0
			p.state = ePath
		case eQuery:
			switch data[i] {
			case ' ':
				p.request.Query.Set(p.startLineBuff[p.offset:])
				p.offset = len(p.startLineBuff)
				p.state = eProto
			case '#':
				p.offset = len(p.startLineBuff)
				p.state = eFragment
			case '%':
				p.state = eQueryDecode1Char
			case '+':
				if uint16(len(p.startLineBuff)) >= p.settings.URL.Length.Maximal {
					return parser.Error, nil, errors.ErrURITooLong
				}

				p.startLineBuff = append(p.startLineBuff, ' ')
			case '\x00', '\n', '\r', '\t', '\b', '\a', '\v', '\f':
				return parser.Error, nil, errors.ErrBadRequest
			default:
				if uint16(len(p.startLineBuff)) >= p.settings.URL.Length.Maximal {
					return parser.Error, nil, errors.ErrURITooLong
				}

				p.startLineBuff = append(p.startLineBuff, data[i])
			}
		case eQueryDecode1Char:
			if !isHex(data[i]) {
				return parser.Error, nil, errors.ErrURIDecoding
			}

			p.urlEncodedChar = unHex(data[i]) << 4
			p.state = eQueryDecode2Char
		case eQueryDecode2Char:
			if !isHex(data[i]) {
				return parser.Error, nil, errors.ErrURIDecoding
			}
			if uint16(len(p.startLineBuff)) >= p.settings.URL.Length.Maximal {
				return parser.Error, nil, errors.ErrURITooLong
			}

			p.startLineBuff = append(p.startLineBuff, p.urlEncodedChar|unHex(data[i]))
			p.urlEncodedChar = 0
			p.state = eQuery
		case eFragment:
			switch data[i] {
			case ' ':
				p.request.Fragment = internal.B2S(p.startLineBuff[p.offset:])
				p.offset = len(p.startLineBuff)
				p.state = eProto
			case '%':
				p.state = eFragmentDecode1Char
			case '\x00', '\n', '\r', '\t', '\b', '\a', '\v', '\f':
				return parser.Error, nil, errors.ErrBadRequest
			default:
				if uint16(len(p.startLineBuff)) >= p.settings.URL.Length.Maximal {
					return parser.Error, nil, errors.ErrURITooLong
				}

				p.startLineBuff = append(p.startLineBuff, data[i])
			}
		case eFragmentDecode1Char:
			if !isHex(data[i]) {
				return parser.Error, nil, errors.ErrURIDecoding
			}

			p.urlEncodedChar = unHex(data[i]) << 4
			p.state = eFragmentDecode2Char
		case eFragmentDecode2Char:
			if !isHex(data[i]) {
				return parser.Error, nil, errors.ErrURIDecoding
			}
			if uint16(len(p.startLineBuff)) >= p.settings.URL.Length.Maximal {
				return parser.Error, nil, errors.ErrURITooLong
			}

			p.startLineBuff = append(p.startLineBuff, p.urlEncodedChar|unHex(data[i]))
			p.urlEncodedChar = 0
			p.state = eFragment
		case eProto:
			switch data[i] {
			case '\r':
				p.state = eProtoCR
			case '\n':
				p.request.Proto = proto.Parse(internal.B2S(p.startLineBuff[p.offset:]))
				if p.request.Proto == proto.Unknown {
					return parser.Error, nil, errors.ErrUnsupportedProtocol
				}
				p.state = eProtoCRLF
			default:
				p.startLineBuff = append(p.startLineBuff, data[i])
				if uint16(len(p.startLineBuff)) >= p.settings.URL.Length.Maximal {
					return parser.Error, nil, errors.ErrUnsupportedProtocol
				}
			}
		case eProtoCR:
			if data[i] != '\n' {
				return parser.Error, nil, errors.ErrBadRequest
			}

			p.state = eProtoCRLF
		case eProtoCRLF:
			p.request.Proto = proto.Parse(internal.B2S(p.startLineBuff[p.offset:]))
			if p.request.Proto == proto.Unknown {
				return parser.Error, nil, errors.ErrUnsupportedProtocol
			}

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
				if p.headersManager.BeginValue() {
					return parser.Error, nil, errors.ErrTooManyHeaders
				}
				p.state = eHeaderColon
			case '\r', '\n':
				return parser.Error, nil, errors.ErrBadRequest
			default:
				if len(p.headerBuff) >= int(p.settings.Headers.KeyLength.Maximal) {
					return parser.Error, nil, errors.ErrHeaderFieldsTooLarge
				}

				p.headerBuff = append(p.headerBuff, data[i]|0x20)
			}
		case eHeaderColon:
			switch data[i] {
			case '\r', '\n':
				return parser.Error, nil, errors.ErrBadRequest
			case ' ':
			default:
				if p.headersManager.AppendValue(data[i]) {
					return parser.Error, nil, errors.ErrHeaderFieldsTooLarge
				}
			}

			p.state = eHeaderValue
		case eHeaderValue:
			switch data[i] {
			case '\r':
				p.state = eHeaderValueCR
			case '\n':
				p.state = eHeaderValueCRLF
			default:
				if p.headersManager.AppendValue(data[i]) {
					return parser.Error, nil, errors.ErrHeaderFieldsTooLarge
				}
			}
		case eHeaderValueCR:
			switch data[i] {
			case '\n':
				p.state = eHeaderValueCRLF
			default:
				return parser.Error, nil, errors.ErrBadRequest
			}
		case eHeaderValueCRLF:
			key := string(p.headerBuff)
			value := p.headersManager.FinalizeValue(key)
			p.request.Headers[key] = value

			switch internal.B2S(p.headerBuff) {
			case "content-length":
				p.contentLength, err = parseUint(value)
				if err != nil {
					return parser.Error, nil, err
				}
				if p.contentLength > uint(p.settings.Body.Length.Maximal) {
					return parser.Error, nil, errors.ErrTooLarge
				}
				p.lengthCountdown = p.contentLength
			case "connection":
				p.closeConnection = string(value) == "close"
			case "transfer-encoding":
				// TODO: parse header value
				p.chunkedTransferEncoding = string(value) == "chunked"
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

// parseBody parses body. In case chunked transfer encoding is active,
// only data chunks will be sent, excluding all the CRLFs and chunk
// lengths
func (p *httpRequestsParser) parseBody(b []byte) (done bool, extra []byte, err error) {
	if p.chunkedTransferEncoding {
		return p.chunkedBodyParser.Parse(b)
	}

	if p.lengthCountdown <= uint(len(b)) {
		p.body.Data <- b[:p.lengthCountdown]
		<-p.body.Data
		extra = b[p.lengthCountdown:]
		p.lengthCountdown = 0

		return true, extra, p.body.Err
	}

	p.body.Data <- b
	<-p.body.Data
	p.lengthCountdown -= uint(len(b))

	return false, nil, p.body.Err
}

// FinalizeBody method just signalizes reader that we're done. This method
// must be called only in 1 case - parser returned parser.RequestCompleted
// state that means headers are parsed, but no body is presented for
// the request, so first starting request processing, then sending a
// completion flag into the body chan
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
