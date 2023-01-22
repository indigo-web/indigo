package http1

import (
	"fmt"

	"github.com/indigo-web/indigo/http"
	"github.com/indigo-web/indigo/http/headers"
	methods "github.com/indigo-web/indigo/http/method"
	"github.com/indigo-web/indigo/http/proto"
	"github.com/indigo-web/indigo/http/status"
	"github.com/indigo-web/indigo/internal"
	"github.com/indigo-web/indigo/internal/alloc"
	"github.com/indigo-web/indigo/internal/parser"
	"github.com/indigo-web/indigo/internal/pool"
	"github.com/indigo-web/indigo/settings"
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
	request *http.Request

	headersSettings settings.Headers

	contentLength           int
	closeConnection         bool
	chunkedTransferEncoding bool
	trailer                 bool

	startLineBuff          []byte
	begin, pointer         int
	urlEncodedChar         uint8
	protoMajor, protoMinor uint8

	headersNumber        int
	headerKey            string
	headerKeyAllocator   alloc.Allocator
	headerValueAllocator alloc.Allocator
	headersValuesPool    pool.ObjectPool[[]string]
}

func NewHTTPRequestsParser(
	request *http.Request, keyAllocator, valAllocator alloc.Allocator,
	valuesPool pool.ObjectPool[[]string], startLineBuff []byte, headersSettings settings.Headers,
) parser.HTTPRequestsParser {
	return &httpRequestsParser{
		state:           eMethod,
		request:         request,
		headersSettings: headersSettings,

		startLineBuff: startLineBuff,

		headerKeyAllocator:   keyAllocator,
		headerValueAllocator: valAllocator,
		headersValuesPool:    valuesPool,
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
	default:
		panic(fmt.Sprintf("BUG: unexpected state: %v", p.state))
	}

method:
	for i := range data {
		switch data[i] {
		case '\r', '\n': // rfc2068, 4.1
			if p.pointer > 0 {
				return parser.Error, nil, status.ErrMethodNotImplemented
			}
		case ' ':
			if p.pointer == 0 {
				return parser.Error, nil, status.ErrBadRequest
			}

			p.request.Method = methods.Parse(internal.B2S(p.startLineBuff[:p.pointer]))

			if p.request.Method == methods.Unknown {
				return parser.Error, nil, status.ErrMethodNotImplemented
			}

			p.begin = p.pointer
			data = data[i+1:]
			p.state = ePath
			goto path
		default:
			if p.pointer > maxMethodLength {
				return parser.Error, nil, status.ErrBadRequest
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
				return parser.Error, nil, status.ErrBadRequest
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
			return parser.Error, nil, status.ErrBadRequest
		default:
			if p.pointer >= len(p.startLineBuff) {
				return parser.Error, nil, status.ErrURITooLong
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
		return parser.Error, nil, status.ErrURIDecoding
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
		return parser.Error, nil, status.ErrURIDecoding
	}

	if p.pointer >= len(p.startLineBuff) {
		return parser.Error, nil, status.ErrURITooLong
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
				return parser.Error, nil, status.ErrURITooLong
			}

			p.startLineBuff[p.pointer] = ' '
			p.pointer++
		case '\x00', '\n', '\r', '\t', '\b', '\a', '\v', '\f':
			return parser.Error, nil, status.ErrBadRequest
		default:
			if p.pointer >= len(p.startLineBuff) {
				return parser.Error, nil, status.ErrURITooLong
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
		return parser.Error, nil, status.ErrURIDecoding
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
		return parser.Error, nil, status.ErrURIDecoding
	}
	if p.pointer >= len(p.startLineBuff) {
		return parser.Error, nil, status.ErrURITooLong
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
			return parser.Error, nil, status.ErrBadRequest
		default:
			if p.pointer >= len(p.startLineBuff) {
				return parser.Error, nil, status.ErrURITooLong
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
		return parser.Error, nil, status.ErrURIDecoding
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
		return parser.Error, nil, status.ErrURIDecoding
	}
	if p.pointer >= len(p.startLineBuff) {
		return parser.Error, nil, status.ErrURITooLong
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
		return parser.Error, nil, status.ErrBadRequest
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
		return parser.Error, nil, status.ErrUnsupportedProtocol
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
		return parser.Error, nil, status.ErrUnsupportedProtocol
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
		return parser.Error, nil, status.ErrUnsupportedProtocol
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
		return parser.Error, nil, status.ErrUnsupportedProtocol
	}

protoMajor:
	if len(data) == 0 {
		return parser.Pending, nil, nil
	}

	if data[0]-'0' > 9 {
		return parser.Error, nil, status.ErrUnsupportedProtocol
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
		return parser.Error, nil, status.ErrUnsupportedProtocol
	}

protoMinor:
	if len(data) == 0 {
		return parser.Pending, nil, nil
	}

	if data[0]-'0' > 9 {
		return parser.Error, nil, status.ErrUnsupportedProtocol
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
		return parser.Error, nil, status.ErrUnsupportedProtocol
	}

protoCR:
	if len(data) == 0 {
		return parser.Pending, nil, nil
	}

	if data[0] != '\n' {
		return parser.Error, nil, status.ErrBadRequest
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
		return parser.Error, nil, status.ErrUnsupportedProtocol
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
		return parser.Error, nil, status.ErrBadRequest
	}

headerKey:
	for i := range data {
		switch data[i] {
		case ':':
			if p.headersNumber > p.headersSettings.Number.Maximal {
				return parser.Error, nil, status.ErrTooManyHeaders
			}

			p.headersNumber++

			if !p.headerKeyAllocator.Append(data[:i]) {
				return parser.Error, nil, status.ErrHeaderFieldsTooLarge
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
			return parser.Error, nil, status.ErrBadRequest
		default:
			data[i] = data[i] | 0x20
		}
	}

	if !p.headerKeyAllocator.Append(data) {
		return parser.Error, nil, status.ErrHeaderFieldsTooLarge
	}

	return parser.Pending, nil, nil

headerColon:
	for i := range data {
		switch data[i] {
		case '\r', '\n':
			return parser.Error, nil, status.ErrBadRequest
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
				return parser.Error, nil, status.ErrBadRequest
			}

			p.contentLength = p.contentLength*10 + int(char-'0')
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
		return parser.Error, nil, status.ErrBadRequest
	}

contentLengthCRLF:
	if len(data) == 0 {
		return parser.Pending, nil, nil
	}

	p.request.ContentLength = p.contentLength

	switch data[0] {
	case '\r':
		data = data[1:]
		p.state = eContentLengthCRLFCR
		goto contentLengthCRLFCR
	case '\n':
		if p.contentLength == 0 && !p.chunkedTransferEncoding {
			return parser.RequestCompleted, data[1:], nil
		}

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
		if p.contentLength == 0 && !p.chunkedTransferEncoding {
			return parser.RequestCompleted, data[1:], nil
		}

		return parser.HeadersCompleted, data[1:], nil
	default:
		return parser.Error, nil, status.ErrBadRequest
	}

headerValue:
	for i := range data {
		switch char := data[i]; char {
		case '\r':
			if !p.headerValueAllocator.Append(data[:i]) {
				return parser.Error, nil, status.ErrHeaderFieldsTooLarge
			}

			data = data[i+1:]
			p.state = eHeaderValueCR
			goto headerValueCR
		case '\n':
			if !p.headerValueAllocator.Append(data[:i]) {
				return parser.Error, nil, status.ErrHeaderFieldsTooLarge
			}

			data = data[i+1:]
			p.state = eHeaderValueCRLF
			goto headerValueCRLF
		}
	}

	if !p.headerValueAllocator.Append(data) {
		return parser.Error, nil, status.ErrHeaderFieldsTooLarge
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
		return parser.Error, nil, status.ErrBadRequest
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
	case "trailer":
		p.trailer = true
	}

	switch data[0] {
	case '\n':
		if p.contentLength == 0 && !p.chunkedTransferEncoding {
			return parser.RequestCompleted, data[1:], nil
		}

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
		if p.contentLength == 0 && !p.chunkedTransferEncoding {
			return parser.RequestCompleted, data[1:], nil
		}

		return parser.HeadersCompleted, data[1:], nil
	default:
		return parser.Error, nil, status.ErrBadRequest
	}
}

func (p *httpRequestsParser) Release() {
	requestHeaders := p.request.Headers.AsMap()

	for _, values := range requestHeaders {
		p.headersValuesPool.Release(values)
	}

	// separated delete-loop from releasing headers values pool because go's standard (Google's)
	// compiler optimizes delete-loop ONLY in case there is a single expression that is a delete
	// function call. So maybe such an optimization will make some difference
	// TODO: profile this and proof that this optimization makes sense
	for k := range requestHeaders {
		delete(requestHeaders, k)
	}

	p.reset()
}

func (p *httpRequestsParser) reset() {
	p.protoMajor = 0
	p.protoMinor = 0
	p.headersNumber = 0
	p.chunkedTransferEncoding = false
	p.headerKeyAllocator.Clear()
	p.headerValueAllocator.Clear()
	p.trailer = false
	p.contentLength = 0
	p.state = eMethod
}
