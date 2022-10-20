package http1

import (
	"github.com/fakefloordiv/indigo/http"
	"github.com/fakefloordiv/indigo/http/encodings"
	methods "github.com/fakefloordiv/indigo/http/method"
	"github.com/fakefloordiv/indigo/http/proto"
	"github.com/fakefloordiv/indigo/internal"
	"github.com/fakefloordiv/indigo/internal/alloc"
	"github.com/fakefloordiv/indigo/internal/body"
	"github.com/fakefloordiv/indigo/internal/parser"
	"github.com/fakefloordiv/indigo/settings"
	"github.com/fakefloordiv/indigo/types"
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

	lengthCountdown         uint
	closeConnection         bool
	chunkedTransferEncoding bool
	trailer                 bool
	chunkedBodyParser       chunkedBodyParser

	startLineBuff          []byte
	offset                 int
	urlEncodedChar         uint8
	protoMajor, protoMinor uint8

	headersNumber        uint8
	headerKey            string
	headerKeyAllocator   alloc.Allocator
	headerValueAllocator alloc.Allocator

	body       *body.Gateway
	codings    encodings.ContentEncodings
	decodeBody bool
	decoder    encodings.Decoder
}

func NewHTTPRequestsParser(
	request *types.Request, body *body.Gateway, keyAllocator, valAllocator alloc.Allocator,
	startLineBuff []byte, settings settings.Settings,
	codings encodings.ContentEncodings,
) parser.HTTPRequestsParser {
	return &httpRequestsParser{
		state:   eMethod,
		request: request,

		chunkedBodyParser:    newChunkedBodyParser(body, settings),
		settings:             settings,
		startLineBuff:        startLineBuff,
		headerKeyAllocator:   keyAllocator,
		headerValueAllocator: valAllocator,
		codings:              codings,

		body: body,
	}
}

func (p *httpRequestsParser) Parse(data []byte) (state parser.RequestState, extra []byte, err error) {
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

	var hBegin int

	for i := range data {
		switch p.state {
		case eMethod:
			switch data[i] {
			case '\r', '\n': // rfc2068, 4.1
				if len(p.startLineBuff) > 0 {
					return parser.Error, nil, http.ErrMethodNotImplemented
				}
			case ' ':
				if len(p.startLineBuff) == 0 {
					return parser.Error, nil, http.ErrBadRequest
				}

				p.request.Method = methods.Parse(internal.B2S(p.startLineBuff))

				if p.request.Method == methods.Unknown {
					return parser.Error, nil, http.ErrMethodNotImplemented
				}

				p.offset = len(p.startLineBuff)
				p.state = ePath
			default:
				if len(p.startLineBuff) > len("CONNECT") { // the longest method, trust me
					return parser.Error, nil, http.ErrBadRequest
				}

				p.startLineBuff = append(p.startLineBuff, data[i])
			}
		case ePath:
			switch data[i] {
			case ' ':
				if len(p.startLineBuff) == p.offset {
					return parser.Error, nil, http.ErrBadRequest
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
				return parser.Error, nil, http.ErrBadRequest
			default:
				if uint16(len(p.startLineBuff)) >= p.settings.URL.Length.Maximal {
					return parser.Error, nil, http.ErrURITooLong
				}

				p.startLineBuff = append(p.startLineBuff, data[i])
			}
		case ePathDecode1Char:
			if !isHex(data[i]) {
				return parser.Error, nil, http.ErrURIDecoding
			}

			p.urlEncodedChar = unHex(data[i]) << 4
			p.state = ePathDecode2Char
		case ePathDecode2Char:
			if !isHex(data[i]) {
				return parser.Error, nil, http.ErrURIDecoding
			}
			if uint16(len(p.startLineBuff)) >= p.settings.URL.Length.Maximal {
				return parser.Error, nil, http.ErrURITooLong
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
					return parser.Error, nil, http.ErrURITooLong
				}

				p.startLineBuff = append(p.startLineBuff, ' ')
			case '\x00', '\n', '\r', '\t', '\b', '\a', '\v', '\f':
				return parser.Error, nil, http.ErrBadRequest
			default:
				if uint16(len(p.startLineBuff)) >= p.settings.URL.Length.Maximal {
					return parser.Error, nil, http.ErrURITooLong
				}

				p.startLineBuff = append(p.startLineBuff, data[i])
			}
		case eQueryDecode1Char:
			if !isHex(data[i]) {
				return parser.Error, nil, http.ErrURIDecoding
			}

			p.urlEncodedChar = unHex(data[i]) << 4
			p.state = eQueryDecode2Char
		case eQueryDecode2Char:
			if !isHex(data[i]) {
				return parser.Error, nil, http.ErrURIDecoding
			}
			if uint16(len(p.startLineBuff)) >= p.settings.URL.Length.Maximal {
				return parser.Error, nil, http.ErrURITooLong
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
				return parser.Error, nil, http.ErrBadRequest
			default:
				if uint16(len(p.startLineBuff)) >= p.settings.URL.Length.Maximal {
					return parser.Error, nil, http.ErrURITooLong
				}

				p.startLineBuff = append(p.startLineBuff, data[i])
			}
		case eFragmentDecode1Char:
			if !isHex(data[i]) {
				return parser.Error, nil, http.ErrURIDecoding
			}

			p.urlEncodedChar = unHex(data[i]) << 4
			p.state = eFragmentDecode2Char
		case eFragmentDecode2Char:
			if !isHex(data[i]) {
				return parser.Error, nil, http.ErrURIDecoding
			}
			if uint16(len(p.startLineBuff)) >= p.settings.URL.Length.Maximal {
				return parser.Error, nil, http.ErrURITooLong
			}

			p.startLineBuff = append(p.startLineBuff, p.urlEncodedChar|unHex(data[i]))
			p.urlEncodedChar = 0
			p.state = eFragment
		case eProto:
			switch data[i] {
			case '\r', '\n':
				return parser.Error, nil, http.ErrBadRequest
			case 'H', 'h':
				p.state = eH
			}
		case eH:
			switch data[i] {
			case 'T', 't':
				p.state = eHT
			default:
				return parser.Error, nil, http.ErrUnsupportedProtocol
			}
		case eHT:
			switch data[i] {
			case 'T', 't':
				p.state = eHTT
			default:
				return parser.Error, nil, http.ErrUnsupportedProtocol
			}
		case eHTT:
			switch data[i] {
			case 'P', 'p':
				p.state = eHTTP
			default:
				return parser.Error, nil, http.ErrUnsupportedProtocol
			}
		case eHTTP:
			switch data[i] {
			case '/':
				p.state = eProtoMajor
			default:
				return parser.Error, nil, http.ErrUnsupportedProtocol
			}
		case eProtoMajor:
			if data[i]-'0' > 9 {
				return parser.Error, nil, http.ErrUnsupportedProtocol
			}

			p.protoMajor = data[i] - '0'
			p.state = eProtoDot
		case eProtoDot:
			switch data[i] {
			case '.':
				p.state = eProtoMinor
			default:
				return parser.Error, nil, http.ErrUnsupportedProtocol
			}
		case eProtoMinor:
			if data[i]-'0' > 9 {
				return parser.Error, nil, http.ErrUnsupportedProtocol
			}

			p.protoMinor = data[i] - '0'
			p.state = eProtoEnd
		case eProtoEnd:
			switch data[i] {
			case '\r':
				p.state = eProtoCR
			case '\n':
				p.state = eProtoCRLF
			default:
				return parser.Error, nil, http.ErrUnsupportedProtocol
			}
		case eProtoCR:
			if data[i] != '\n' {
				return parser.Error, nil, http.ErrBadRequest
			}

			p.state = eProtoCRLF
		case eProtoCRLF:
			p.request.Proto = proto.Parse(p.protoMajor, p.protoMinor)
			if p.request.Proto == proto.Unknown {
				return parser.Error, nil, http.ErrUnsupportedProtocol
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
				hBegin = i
				data[i] = data[i] | 0x20
				p.state = eHeaderKey
			}
		case eProtoCRLFCR:
			switch data[i] {
			case '\n':
				// no request body because even no headers
				p.reset()

				return parser.RequestCompleted, data[i+1:], nil
			default:
				return parser.Error, nil, http.ErrBadRequest
			}
		case eHeaderKey:
			switch data[i] {
			case ':':
				if p.headersNumber > p.settings.Headers.Number.Maximal {
					return parser.Error, nil, http.ErrTooManyHeaders
				}

				p.headersNumber++

				if !p.headerKeyAllocator.Append(data[hBegin:i]) {
					return parser.Error, nil, http.ErrHeaderFieldsTooLarge
				}

				p.headerKey = internal.B2S(p.headerKeyAllocator.Finish())

				if p.headerKey == "content-length" {
					p.state = eContentLength
					continue
				}

				p.state = eHeaderColon
			case '\r', '\n':
				return parser.Error, nil, http.ErrBadRequest
			default:
				data[i] = data[i] | 0x20
			}
		case eHeaderColon:
			switch data[i] {
			case '\r', '\n':
				return parser.Error, nil, http.ErrBadRequest
			case ' ':
			case '\\':
				hBegin = i
				p.state = eHeaderValueBackslash
			default:
				hBegin = i
				p.state = eHeaderValue
			}
		case eContentLength:
			switch char := data[i]; char {
			case ' ':
			case '\r':
				p.state = eContentLengthCR
			case '\n':
				p.state = eContentLengthCRLF
			default:
				if char < '0' || char > '9' {
					return parser.Error, nil, http.ErrBadRequest
				}

				p.lengthCountdown = p.lengthCountdown*10 + uint(char-'0')
			}
		case eContentLengthCR:
			switch data[i] {
			case '\n':
				p.state = eContentLengthCRLF
			default:
				return parser.Error, nil, http.ErrBadRequest
			}
		case eContentLengthCRLF:
			p.request.ContentLength = p.lengthCountdown

			switch data[i] {
			case '\r':
				p.state = eContentLengthCRLFCR
			case '\n':
				if p.lengthCountdown == 0 && !p.chunkedTransferEncoding {
					p.reset()

					return parser.RequestCompleted, data[i+1:], nil
				}

				p.state = eBody

				return parser.HeadersCompleted, data[i+1:], nil
			default:
				hBegin = i
				data[i] = data[i] | 0x20
				p.state = eHeaderKey
			}
		case eContentLengthCRLFCR:
			switch data[i] {
			case '\n':
				if p.lengthCountdown == 0 && !p.chunkedTransferEncoding {
					p.reset()

					return parser.RequestCompleted, data[i+1:], nil
				}

				p.state = eBody

				return parser.HeadersCompleted, data[i+1:], nil
			default:
				return parser.Error, nil, http.ErrBadRequest
			}
		case eHeaderValue:
			switch char := data[i]; char {
			case '\r':
				if !p.headerValueAllocator.Append(data[hBegin:i]) {
					return parser.Error, nil, http.ErrHeaderFieldsTooLarge
				}

				p.state = eHeaderValueCR
			case '\n':
				if !p.headerValueAllocator.Append(data[hBegin:i]) {
					return parser.Error, nil, http.ErrHeaderFieldsTooLarge
				}

				p.state = eHeaderValueCRLF
			case '\\':
				p.state = eHeaderValueBackslash
			case '"':
				p.state = eHeaderValueQuoted
			case ',':
				// When comma is met, current header is finalized, and a new one started with the same key
				// In case it is a system header like Content-Length, Connection, etc., we just ignore it
				// because they anyway must not include commas in their values
				p.state = eHeaderValueComma

				if !p.headerValueAllocator.Append(data[hBegin:i]) {
					return parser.Error, nil, http.ErrHeaderFieldsTooLarge
				}

				value := p.headerValueAllocator.Finish()
				p.request.Headers.Add(p.headerKey, internal.B2S(value))
			}
		case eHeaderValueComma:
			switch char := data[i]; char {
			case ' ':
			case '\r':
				if !p.headerValueAllocator.Append(data[hBegin:i]) {
					return parser.Error, nil, http.ErrHeaderFieldsTooLarge
				}

				p.state = eHeaderValueCR
			case '\n':
				if !p.headerValueAllocator.Append(data[hBegin:i]) {
					return parser.Error, nil, http.ErrHeaderFieldsTooLarge
				}

				p.state = eHeaderValueCRLF
			case '\\':
				hBegin = i
				p.state = eHeaderValueBackslash
			case '"':
				hBegin = i
				p.state = eHeaderValueQuoted
			default:
				hBegin = i
				p.state = eHeaderValue
			}
		case eHeaderValueQuoted:
			switch char := data[i]; char {
			case '\r':
				if !p.headerValueAllocator.Append(data[hBegin:i]) {
					return parser.Error, nil, http.ErrHeaderFieldsTooLarge
				}

				p.state = eHeaderValueCR
			case '\n':
				if !p.headerValueAllocator.Append(data[hBegin:i]) {
					return parser.Error, nil, http.ErrHeaderFieldsTooLarge
				}

				p.state = eHeaderValueCRLF
			case '\\':
				p.state = eHeaderValueQuotedBackslash
			case '"':
				p.state = eHeaderValue
			}
		case eHeaderValueBackslash:
			p.state = eHeaderValue
		case eHeaderValueQuotedBackslash:
			p.state = eHeaderValueQuoted
		case eHeaderValueCR:
			switch data[i] {
			case '\n':
				p.state = eHeaderValueCRLF
			default:
				return parser.Error, nil, http.ErrBadRequest
			}
		case eHeaderValueCRLF:
			// TODO: save whole header line, without any commas, here. Only after that
			//       analyze how long our slice is supposed to be
			value := p.headerValueAllocator.Finish()
			p.request.Headers.Add(p.headerKey, internal.B2S(value))

			switch p.headerKey {
			case "connection":
				p.closeConnection = string(value) == "close"
			case "transfer-encoding":
				p.chunkedTransferEncoding = string(value) == "chunked"
			case "content-encoding":
				p.decoder, p.decodeBody = p.codings.GetDecoder(internal.B2S(value))
				if !p.decodeBody {
					return parser.Error, nil, http.ErrUnsupportedEncoding
				}
			case "trailer":
				p.trailer = true
			}

			switch data[i] {
			case '\n':
				if p.lengthCountdown == 0 && !p.chunkedTransferEncoding {
					p.reset()

					return parser.RequestCompleted, data[i+1:], nil
				}

				p.state = eBody

				return parser.HeadersCompleted, data[i+1:], nil
			case '\r':
				p.state = eHeaderValueCRLFCR
			default:
				hBegin = i
				data[i] = data[i] | 0x20
				p.state = eHeaderKey
			}
		case eHeaderValueCRLFCR:
			switch data[i] {
			case '\n':
				if p.lengthCountdown == 0 && !p.chunkedTransferEncoding {
					p.reset()

					return parser.RequestCompleted, data[i+1:], nil
				}

				p.state = eBody

				return parser.HeadersCompleted, data[i+1:], nil
			default:
				return parser.Error, nil, http.ErrBadRequest
			}
		}
	}

	switch p.state {
	case eHeaderValue:
		if !p.headerValueAllocator.Append(data[hBegin:]) {
			return parser.Error, nil, http.ErrHeaderFieldsTooLarge
		}
	case eHeaderKey:
		if !p.headerKeyAllocator.Append(data[hBegin:]) {
			return parser.Error, nil, http.ErrHeaderFieldsTooLarge
		}
	}

	return parser.Pending, nil, nil
}

// parseBody parses body. In case chunked transfer encoding is active,
// only data chunks will be sent, excluding all the CRLFs and chunk
// lengths
func (p *httpRequestsParser) parseBody(b []byte) (done bool, extra []byte, err error) {
	if p.chunkedTransferEncoding {
		return p.chunkedBodyParser.Parse(b, p.decoder, p.trailer)
	}

	if p.lengthCountdown <= uint(len(b)) {
		piece := b[:p.lengthCountdown]
		if p.decodeBody {
			piece, err = p.decoder(piece)
			if err != nil {
				return true, nil, err
			}
		}

		p.body.Data <- piece
		<-p.body.Data
		extra = b[p.lengthCountdown:]
		p.lengthCountdown = 0

		return true, extra, p.body.Err
	}

	piece := b
	if p.decodeBody {
		piece, err = p.decoder(b)
		if err != nil {
			return true, nil, err
		}
	}

	p.body.Data <- piece
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
	p.protoMajor = 0
	p.protoMinor = 0
	p.offset = 0
	p.headersNumber = 0
	p.chunkedTransferEncoding = false
	p.decodeBody = false
	p.headerKeyAllocator.Clear()
	p.headerValueAllocator.Clear()
	p.trailer = false
	p.state = eMethod
}
