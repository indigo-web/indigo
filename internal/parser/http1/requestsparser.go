package http1

import (
	"github.com/fakefloordiv/indigo/http"
	"github.com/fakefloordiv/indigo/http/encodings"
	"github.com/fakefloordiv/indigo/http/headers"
	methods "github.com/fakefloordiv/indigo/http/method"
	"github.com/fakefloordiv/indigo/http/proto"
	"github.com/fakefloordiv/indigo/internal"
	"github.com/fakefloordiv/indigo/internal/alloc"
	"github.com/fakefloordiv/indigo/internal/body"
	"github.com/fakefloordiv/indigo/internal/parser"
	"github.com/fakefloordiv/indigo/internal/pool"
	"github.com/fakefloordiv/indigo/settings"
	"github.com/fakefloordiv/indigo/types"
)

const maxMethodLength = len("CONNECT")

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
	begin, pointer         int
	urlEncodedChar         uint8
	protoMajor, protoMinor uint8

	headersNumber        uint8
	headerKey            string
	headerKeyAllocator   alloc.Allocator
	headerValueAllocator alloc.Allocator
	headersValuesPool    pool.ObjectPool[[]string]

	body       *body.Gateway
	decoders   encodings.Decoders
	decoder    encodings.DecoderFunc
	decodeBody bool
}

func NewHTTPRequestsParser(
	request *types.Request, body *body.Gateway, keyAllocator, valAllocator alloc.Allocator,
	valuesPool pool.ObjectPool[[]string], startLineBuff []byte, settings settings.Settings,
	decoders encodings.Decoders,
) parser.HTTPRequestsParser {
	return &httpRequestsParser{
		state:    eMethod,
		request:  request,
		settings: settings,

		chunkedBodyParser: newChunkedBodyParser(body, settings),
		startLineBuff:     startLineBuff,

		headerKeyAllocator:   keyAllocator,
		headerValueAllocator: valAllocator,
		headersValuesPool:    valuesPool,

		body:     body,
		decoders: decoders,
	}
}

func (p *httpRequestsParser) Parse(data []byte) (state parser.RequestState, extra []byte, err error) {
	var value string
	requestHeaders := p.request.Headers

	switch p.state {
	case eMethod:
		goto method
	case ePath:
		goto path
	case ePathDecode1Char:
		goto pathDecode1Char
	case ePathDecode2Char:
		goto pathDecode2Char
	case eQuery:
		goto query
	case eQueryDecode1Char:
		goto queryDecode1Char
	case eQueryDecode2Char:
		goto queryDecode2Char
	case eFragment:
		goto fragment
	case eFragmentDecode1Char:
		goto fragmentDecode1Char
	case eFragmentDecode2Char:
		goto fragmentDecode2Char
	case eProto:
		goto proto
	case eH:
		goto protoH
	case eHT:
		goto protoHT
	case eHTT:
		goto protoHTT
	case eHTTP:
		goto protoHTTP
	case eProtoMajor:
		goto protoMajor
	case eProtoDot:
		goto protoDot
	case eProtoMinor:
		goto protoMinor
	case eProtoEnd:
		goto protoEnd
	case eProtoCR:
		goto protoCR
	case eProtoCRLF:
		goto protoCRLF
	case eProtoCRLFCR:
		goto protoCRLFCR
	case eHeaderKey:
		goto headerKey
	case eHeaderColon:
		goto headerColon
	case eContentLength:
		goto contentLength
	case eContentLengthCR:
		goto contentLengthCR
	case eContentLengthCRLF:
		goto contentLengthCRLF
	case eContentLengthCRLFCR:
		goto contentLengthCRLFCR
	case eHeaderValue:
		goto headerValue
	case eHeaderValueCR:
		goto headerValueCR
	case eHeaderValueCRLF:
		goto headerValueCRLF
	case eHeaderValueCRLFCR:
		goto headerValueCRLFCR
	case eBody:
		if p.chunkedTransferEncoding {
			done, extra, err := p.chunkedBodyParser.Parse(data, p.decoder, p.trailer)
			if err != nil {
				p.body.WriteErr(err)

				return parser.Error, nil, err
			} else if done {
				p.body.Data <- nil

				return parser.BodyCompleted, extra, nil
			}

			return parser.Pending, extra, nil
		}

		if p.lengthCountdown <= uint(len(data)) {
			piece := data[:p.lengthCountdown]
			if p.decodeBody {
				piece, err = p.decoder(piece)
				if err != nil {
					p.body.WriteErr(err)

					return parser.Error, nil, err
				}
			}

			p.body.Data <- piece
			<-p.body.Data
			p.body.Data <- nil
			extra = data[p.lengthCountdown:]
			p.lengthCountdown = 0

			if p.body.Err != nil {
				return parser.Error, extra, p.body.Err
			}

			return parser.BodyCompleted, extra, nil
		}

		piece := data
		if p.decodeBody {
			piece, err = p.decoder(data)
			if err != nil {
				p.body.WriteErr(err)

				return parser.Error, nil, err
			}
		}

		p.body.Data <- piece
		<-p.body.Data
		p.lengthCountdown -= uint(len(data))

		if p.body.Err != nil {
			return parser.Error, nil, p.body.Err
		}

		return parser.Pending, nil, nil
	}

method:
	for i := range data {
		switch data[i] {
		case '\r', '\n': // rfc2068, 4.1
			if p.pointer > 0 {
				return parser.Error, nil, http.ErrMethodNotImplemented
			}
		case ' ':
			if p.pointer == 0 {
				return parser.Error, nil, http.ErrBadRequest
			}

			p.request.Method = methods.Parse(internal.B2S(p.startLineBuff[:p.pointer]))

			if p.request.Method == methods.Unknown {
				return parser.Error, nil, http.ErrMethodNotImplemented
			}

			p.begin = p.pointer
			data = data[i+1:]
			p.state = ePath
			goto path
		default:
			if p.pointer > maxMethodLength {
				return parser.Error, nil, http.ErrBadRequest
			}

			p.startLineBuff[p.pointer] = data[i]
			p.pointer++
		}
	}

	return parser.Pending, nil, nil

path:
	for i := range data {
		switch data[i] {
		case ' ':
			if p.begin == p.pointer {
				return parser.Error, nil, http.ErrBadRequest
			}

			p.request.Path = internal.B2S(p.startLineBuff[p.begin:p.pointer])
			data = data[i+1:]
			p.state = eProto
			goto proto
		case '%':
			data = data[i+1:]
			p.state = ePathDecode1Char
			goto pathDecode1Char
		case '?':
			p.request.Path = internal.B2S(p.startLineBuff[p.begin:p.pointer])
			if len(p.request.Path) == 0 {
				p.request.Path = "/"
			}

			p.begin = p.pointer
			data = data[i+1:]
			p.state = eQuery
			goto query
		case '#':
			p.request.Path = internal.B2S(p.startLineBuff[p.begin:p.pointer])
			if len(p.request.Path) == 0 {
				p.request.Path = "/"
			}

			p.begin = p.pointer
			data = data[i+1:]
			p.state = eFragment
			goto fragment
		case '\x00', '\n', '\r', '\t', '\b', '\a', '\v', '\f':
			// request path MUST NOT include any non-printable characters
			return parser.Error, nil, http.ErrBadRequest
		default:
			if p.pointer >= len(p.startLineBuff) {
				return parser.Error, nil, http.ErrURITooLong
			}

			p.startLineBuff[p.pointer] = data[i]
			p.pointer++
		}
	}

	return parser.Pending, nil, nil

pathDecode1Char:
	if len(data) == 0 {
		return parser.Pending, nil, nil
	}

	if !isHex(data[0]) {
		return parser.Error, nil, http.ErrURIDecoding
	}

	p.urlEncodedChar = unHex(data[0]) << 4
	data = data[1:]
	p.state = ePathDecode2Char
	goto pathDecode2Char

pathDecode2Char:
	if len(data) == 0 {
		return parser.Pending, nil, nil
	}

	if !isHex(data[0]) {
		return parser.Error, nil, http.ErrURIDecoding
	}

	if p.pointer >= len(p.startLineBuff) {
		return parser.Error, nil, http.ErrURITooLong
	}

	p.startLineBuff[p.pointer] = p.urlEncodedChar | unHex(data[0])
	p.pointer++
	data = data[1:]
	p.state = ePath
	goto path

query:
	for i := range data {
		switch data[i] {
		case ' ':
			p.request.Query.Set(p.startLineBuff[p.begin:p.pointer])
			data = data[i+1:]
			p.state = eProto
			goto proto
		case '#':
			p.begin = p.pointer
			data = data[i+1:]
			p.state = eFragment
			goto fragment
		case '%':
			data = data[i+1:]
			p.state = eQueryDecode1Char
			goto queryDecode1Char
		case '+':
			if p.pointer >= len(p.startLineBuff) {
				return parser.Error, nil, http.ErrURITooLong
			}

			p.startLineBuff[p.pointer] = ' '
			p.pointer++
		case '\x00', '\n', '\r', '\t', '\b', '\a', '\v', '\f':
			return parser.Error, nil, http.ErrBadRequest
		default:
			if p.pointer >= len(p.startLineBuff) {
				return parser.Error, nil, http.ErrURITooLong
			}

			p.startLineBuff[p.pointer] = data[i]
			p.pointer++
		}
	}

	return parser.Pending, nil, nil

queryDecode1Char:
	if len(data) == 0 {
		return parser.Pending, nil, nil
	}

	if !isHex(data[0]) {
		return parser.Error, nil, http.ErrURIDecoding
	}

	p.urlEncodedChar = unHex(data[0]) << 4
	data = data[1:]
	p.state = eQueryDecode2Char
	goto queryDecode2Char

queryDecode2Char:
	if len(data) == 0 {
		return parser.Pending, nil, nil
	}

	if !isHex(data[0]) {
		return parser.Error, nil, http.ErrURIDecoding
	}
	if p.pointer >= len(p.startLineBuff) {
		return parser.Error, nil, http.ErrURITooLong
	}

	p.startLineBuff[p.pointer] = p.urlEncodedChar | unHex(data[0])
	p.pointer++
	data = data[1:]
	p.state = eQuery
	goto query

fragment:
	for i := range data {
		switch data[i] {
		case ' ':
			p.request.Fragment = internal.B2S(p.startLineBuff[p.begin:p.pointer])
			data = data[i+1:]
			p.state = eProto
			goto proto
		case '%':
			data = data[i+1:]
			p.state = eFragmentDecode1Char
			goto fragmentDecode1Char
		case '\x00', '\n', '\r', '\t', '\b', '\a', '\v', '\f':
			return parser.Error, nil, http.ErrBadRequest
		default:
			if p.pointer >= len(p.startLineBuff) {
				return parser.Error, nil, http.ErrURITooLong
			}

			p.startLineBuff[p.pointer] = data[i]
			p.pointer++
		}
	}

	return parser.Pending, nil, nil

fragmentDecode1Char:
	if len(data) == 0 {
		return parser.Pending, nil, nil
	}

	if !isHex(data[0]) {
		return parser.Error, nil, http.ErrURIDecoding
	}

	p.urlEncodedChar = unHex(data[0]) << 4
	data = data[1:]
	p.state = eFragmentDecode2Char
	goto fragmentDecode2Char

fragmentDecode2Char:
	if len(data) == 0 {
		return parser.Pending, nil, nil
	}

	if !isHex(data[0]) {
		return parser.Error, nil, http.ErrURIDecoding
	}
	if p.pointer >= len(p.startLineBuff) {
		return parser.Error, nil, http.ErrURITooLong
	}

	p.startLineBuff[p.pointer] = p.urlEncodedChar | unHex(data[0])
	p.pointer++
	data = data[1:]
	p.state = eFragment
	goto fragment

proto:
	if len(data) == 0 {
		return parser.Pending, nil, nil
	}

	switch data[0] {
	case 'H', 'h':
		p.begin = 0
		p.pointer = 0
		data = data[1:]
		p.state = eH
		goto protoH
	case '\r', '\n':
		return parser.Error, nil, http.ErrBadRequest
	}

protoH:
	if len(data) == 0 {
		return parser.Pending, nil, nil
	}

	switch data[0] {
	case 'T', 't':
		data = data[1:]
		p.state = eHT
		goto protoHT
	default:
		return parser.Error, nil, http.ErrUnsupportedProtocol
	}

protoHT:
	if len(data) == 0 {
		return parser.Pending, nil, nil
	}

	switch data[0] {
	case 'T', 't':
		data = data[1:]
		p.state = eHTT
		goto protoHTT
	default:
		return parser.Error, nil, http.ErrUnsupportedProtocol
	}

protoHTT:
	if len(data) == 0 {
		return parser.Pending, nil, nil
	}

	switch data[0] {
	case 'P', 'p':
		data = data[1:]
		p.state = eHTTP
		goto protoHTTP
	default:
		return parser.Error, nil, http.ErrUnsupportedProtocol
	}

protoHTTP:
	if len(data) == 0 {
		return parser.Pending, nil, nil
	}

	switch data[0] {
	case '/':
		data = data[1:]
		p.state = eProtoMajor
		goto protoMajor
	default:
		return parser.Error, nil, http.ErrUnsupportedProtocol
	}

protoMajor:
	if len(data) == 0 {
		return parser.Pending, nil, nil
	}

	if data[0]-'0' > 9 {
		return parser.Error, nil, http.ErrUnsupportedProtocol
	}

	p.protoMajor = data[0] - '0'
	data = data[1:]
	p.state = eProtoDot
	goto protoDot

protoDot:
	if len(data) == 0 {
		return parser.Pending, nil, nil
	}

	switch data[0] {
	case '.':
		data = data[1:]
		p.state = eProtoMinor
		goto protoMinor
	default:
		return parser.Error, nil, http.ErrUnsupportedProtocol
	}

protoMinor:
	if len(data) == 0 {
		return parser.Pending, nil, nil
	}

	if data[0]-'0' > 9 {
		return parser.Error, nil, http.ErrUnsupportedProtocol
	}

	p.protoMinor = data[0] - '0'
	data = data[1:]
	p.state = eProtoEnd
	goto protoEnd

protoEnd:
	if len(data) == 0 {
		return parser.Pending, nil, nil
	}

	switch data[0] {
	case '\r':
		data = data[1:]
		p.state = eProtoCR
		goto protoCR
	case '\n':
		data = data[1:]
		p.state = eProtoCRLF
		goto protoCRLF
	default:
		return parser.Error, nil, http.ErrUnsupportedProtocol
	}

protoCR:
	if len(data) == 0 {
		return parser.Pending, nil, nil
	}

	if data[0] != '\n' {
		return parser.Error, nil, http.ErrBadRequest
	}

	data = data[1:]
	p.state = eProtoCRLF
	goto protoCRLF

protoCRLF:
	if len(data) == 0 {
		return parser.Pending, nil, nil
	}

	p.request.Proto = proto.Parse(p.protoMajor, p.protoMinor)
	if p.request.Proto == proto.Unknown {
		return parser.Error, nil, http.ErrUnsupportedProtocol
	}

	switch data[0] {
	case '\r':
		data = data[1:]
		p.state = eProtoCRLFCR
		goto protoCRLFCR
	case '\n':
		return parser.RequestCompleted, data[1:], nil
	default:
		// headers are here. I have to have a buffer for header key, and after receiving it,
		// get an appender from headers manager (and keep it in httpRequestsParser struct)
		data[0] = data[0] | 0x20
		p.state = eHeaderKey
		goto headerKey
	}

protoCRLFCR:
	if len(data) == 0 {
		return parser.Pending, nil, nil
	}

	switch data[0] {
	case '\n':
		// no request body because even no headers
		return parser.RequestCompleted, data[1:], nil
	default:
		return parser.Error, nil, http.ErrBadRequest
	}

headerKey:
	for i := range data {
		switch data[i] {
		case ':':
			if p.headersNumber > p.settings.Headers.Number.Maximal {
				return parser.Error, nil, http.ErrTooManyHeaders
			}

			p.headersNumber++

			if !p.headerKeyAllocator.Append(data[:i]) {
				return parser.Error, nil, http.ErrHeaderFieldsTooLarge
			}

			p.headerKey = internal.B2S(p.headerKeyAllocator.Finish())
			data = data[i+1:]

			if p.headerKey == "content-length" {
				p.state = eContentLength
				goto contentLength
			}

			p.state = eHeaderColon
			goto headerColon
		case '\r', '\n':
			return parser.Error, nil, http.ErrBadRequest
		default:
			data[i] = data[i] | 0x20
		}
	}

	if !p.headerKeyAllocator.Append(data) {
		return parser.Error, nil, http.ErrHeaderFieldsTooLarge
	}

	return parser.Pending, nil, nil

headerColon:
	for i := range data {
		switch data[i] {
		case '\r', '\n':
			return parser.Error, nil, http.ErrBadRequest
		case ' ':
		default:
			data = data[i:]
			p.state = eHeaderValue
			goto headerValue
		}
	}

	return parser.Pending, nil, nil

contentLength:
	for i := range data {
		switch char := data[i]; char {
		case ' ':
		case '\r':
			data = data[i+1:]
			p.state = eContentLengthCR
			goto contentLengthCR
		case '\n':
			data = data[i+1:]
			p.state = eContentLengthCRLF
			goto contentLengthCRLF
		default:
			if char < '0' || char > '9' {
				return parser.Error, nil, http.ErrBadRequest
			}

			p.lengthCountdown = p.lengthCountdown*10 + uint(char-'0')
		}
	}

	return parser.Pending, nil, nil

contentLengthCR:
	if len(data) == 0 {
		return parser.Pending, nil, nil
	}

	switch data[0] {
	case '\n':
		data = data[1:]
		p.state = eContentLengthCRLF
		goto contentLengthCRLF
	default:
		return parser.Error, nil, http.ErrBadRequest
	}

contentLengthCRLF:
	if len(data) == 0 {
		return parser.Pending, nil, nil
	}

	p.request.ContentLength = p.lengthCountdown

	switch data[0] {
	case '\r':
		data = data[1:]
		p.state = eContentLengthCRLFCR
		goto contentLengthCRLFCR
	case '\n':
		if p.lengthCountdown == 0 && !p.chunkedTransferEncoding {
			return parser.RequestCompleted, data[1:], nil
		}

		p.state = eBody

		return parser.HeadersCompleted, data[1:], nil
	default:
		data[0] = data[0] | 0x20
		p.state = eHeaderKey
		goto headerKey
	}

contentLengthCRLFCR:
	if len(data) == 0 {
		return parser.Pending, nil, nil
	}

	switch data[0] {
	case '\n':
		if p.lengthCountdown == 0 && !p.chunkedTransferEncoding {
			return parser.RequestCompleted, data[1:], nil
		}

		p.state = eBody

		return parser.HeadersCompleted, data[1:], nil
	default:
		return parser.Error, nil, http.ErrBadRequest
	}

headerValue:
	for i := range data {
		switch char := data[i]; char {
		case '\r':
			if !p.headerValueAllocator.Append(data[:i]) {
				return parser.Error, nil, http.ErrHeaderFieldsTooLarge
			}

			data = data[i+1:]
			p.state = eHeaderValueCR
			goto headerValueCR
		case '\n':
			if !p.headerValueAllocator.Append(data[:i]) {
				return parser.Error, nil, http.ErrHeaderFieldsTooLarge
			}

			data = data[i+1:]
			p.state = eHeaderValueCRLF
			goto headerValueCRLF
		}
	}

	if !p.headerValueAllocator.Append(data) {
		return parser.Error, nil, http.ErrHeaderFieldsTooLarge
	}

	return parser.Pending, nil, nil

headerValueCR:
	if len(data) == 0 {
		return parser.Pending, nil, nil
	}

	switch data[0] {
	case '\n':
		data = data[1:]
		p.state = eHeaderValueCRLF
		goto headerValueCRLF
	default:
		return parser.Error, nil, http.ErrBadRequest
	}

headerValueCRLF:
	if len(data) == 0 {
		return parser.Pending, nil, nil
	}

	value = internal.B2S(p.headerValueAllocator.Finish())

	if buff := requestHeaders.Values(p.headerKey); buff != nil {
		requestHeaders.Add(p.headerKey, value)
	} else {
		buff = p.headersValuesPool.Acquire()[:0]
		buff = append(buff, value)
		requestHeaders.Set(p.headerKey, buff)
	}

	switch p.headerKey {
	case "connection":
		p.closeConnection = value == "close"
	case "transfer-encoding":
		p.chunkedTransferEncoding = headers.ValueOf(value) == "chunked"
		p.request.ChunkedTE = p.chunkedTransferEncoding
	case "content-encoding":
		decoder, found := p.decoders.Get(value)
		if !found {
			return parser.Error, nil, http.ErrUnsupportedEncoding
		}

		p.decodeBody = true
		p.decoder = decoder.New()
	case "trailer":
		p.trailer = true
	}

	switch data[0] {
	case '\n':
		if p.lengthCountdown == 0 && !p.chunkedTransferEncoding {
			return parser.RequestCompleted, data[1:], nil
		}

		p.state = eBody

		return parser.HeadersCompleted, data[1:], nil
	case '\r':
		data = data[1:]
		p.state = eHeaderValueCRLFCR
		goto headerValueCRLFCR
	default:
		data[0] = data[0] | 0x20
		p.state = eHeaderKey
		goto headerKey
	}

headerValueCRLFCR:
	if len(data) == 0 {
		return parser.Pending, nil, nil
	}

	switch data[0] {
	case '\n':
		if p.lengthCountdown == 0 && !p.chunkedTransferEncoding {
			return parser.RequestCompleted, data[1:], nil
		}

		p.state = eBody

		return parser.HeadersCompleted, data[1:], nil
	default:
		return parser.Error, nil, http.ErrBadRequest
	}
}

func (p *httpRequestsParser) Release() {
	requestHeaders := p.request.Headers.AsMap()

	for key, values := range requestHeaders {
		p.headersValuesPool.Release(values)
		delete(requestHeaders, key)
	}

	p.reset()
}

func (p *httpRequestsParser) Release() {
	requestHeaders := p.request.Headers.AsMap()

	for key, values := range requestHeaders {
		p.headersValuesPool.Release(values)
		delete(requestHeaders, key)
	}

	p.reset()
}

func (p *httpRequestsParser) reset() {
	p.protoMajor = 0
	p.protoMinor = 0
	p.headersNumber = 0
	p.chunkedTransferEncoding = false
	p.decodeBody = false
	p.headerKeyAllocator.Clear()
	p.headerValueAllocator.Clear()
	p.trailer = false
	p.state = eMethod
}
