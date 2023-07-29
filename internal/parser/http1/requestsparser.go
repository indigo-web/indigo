package http1

import (
	"bytes"
	"fmt"
	"github.com/indigo-web/indigo/internal/strcomp"
	"strings"

	"github.com/indigo-web/indigo/http"
	"github.com/indigo-web/indigo/http/headers"
	"github.com/indigo-web/indigo/http/method"
	"github.com/indigo-web/indigo/http/proto"
	"github.com/indigo-web/indigo/http/status"
	"github.com/indigo-web/indigo/internal/parser"
	"github.com/indigo-web/indigo/settings"
	"github.com/indigo-web/utils/arena"
	"github.com/indigo-web/utils/pool"
	"github.com/indigo-web/utils/uf"
)

const maxMethodLength = len("CONNECT")

// httpRequestsParser is a stream-based http requests parser. It modifies
// request object by pointer in performance purposes. Decodes query-encoded
// values by its own, you can see that by presented states ePathDecode1Char,
// ePathDecode2Char, etc. When headers are parsed, parser returns state
// parser.HeadersCompleted to notify http server about this, attaching all
// the pending data as an extra. Body must be processed separately
type httpRequestsParser struct {
	request           *http.Request
	headerKey         string
	startLineBuff     []byte
	headersValuesPool pool.ObjectPool[[]string]
	headerKeyArena    arena.Arena[byte]
	headerValueArena  arena.Arena[byte]
	headersSettings   settings.Headers
	begin             int
	pointer           int
	headersNumber     int
	contentLength     int
	headerValueSize   int
	urlEncodedChar    uint8
	protoMajor        uint8
	protoMinor        uint8
	state             parserState
}

func NewHTTPRequestsParser(
	request *http.Request, keyArena, valArena arena.Arena[byte],
	valuesPool pool.ObjectPool[[]string], startLineBuff []byte, headersSettings settings.Headers,
) parser.HTTPRequestsParser {
	return &httpRequestsParser{
		state:           eMethod,
		request:         request,
		headersSettings: headersSettings,

		startLineBuff: startLineBuff,

		headerKeyArena:    keyArena,
		headerValueArena:  valArena,
		headersValuesPool: valuesPool,
	}
}

func (p *httpRequestsParser) Parse(data []byte) (state parser.RequestState, extra []byte, err error) {
	_ = *p.request
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
	case eContentLength:
		goto contentLength
	case eContentLengthCR:
		goto contentLengthCR
	case eContentLengthCRLFCR:
		goto contentLengthCRLFCR
	case eHeaderValue:
		goto headerValue
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

			p.request.Method = method.Parse(uf.B2S(p.startLineBuff[:p.pointer]))

			if p.request.Method == method.Unknown {
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

			p.request.Path.String = uf.B2S(p.startLineBuff[p.begin:p.pointer])
			data = data[i+1:]
			p.state = eProto
			goto proto
		case '%':
			data = data[i+1:]
			p.state = ePathDecode1Char
			goto pathDecode1Char
		case '?':
			p.request.Path.String = uf.B2S(p.startLineBuff[p.begin:p.pointer])
			if len(p.request.Path.String) == 0 {
				p.request.Path.String = "/"
			}

			p.begin = p.pointer
			data = data[i+1:]
			p.state = eQuery
			goto query
		case '#':
			p.request.Path.String = uf.B2S(p.startLineBuff[p.begin:p.pointer])
			if len(p.request.Path.String) == 0 {
				p.request.Path.String = "/"
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
			p.request.Path.Query.Set(p.startLineBuff[p.begin:p.pointer])
			data = data[i+1:]
			p.state = eProto
			goto proto
		case '#':
			p.request.Path.Query.Set(p.startLineBuff[p.begin:p.pointer])
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
	{
		sp := bytes.IndexByte(data, ' ')
		if sp == -1 {
			return parser.Pending, nil, nil
		}

		data = data[sp+1:]
		p.state = eProto
		goto proto
	}

proto:
	if len(data) == 0 {
		return parser.Pending, nil, nil
	}

	if data[0]|0x20 == 'h' {
		data = data[1:]
		p.state = eH
		goto protoH
	}

	return parser.Error, nil, status.ErrBadRequest

protoH:
	if len(data) == 0 {
		return parser.Pending, nil, nil
	}

	if data[0]|0x20 == 't' {
		data = data[1:]
		p.state = eHT
		goto protoHT
	}

	return parser.Error, nil, status.ErrUnsupportedProtocol

protoHT:
	if len(data) == 0 {
		return parser.Pending, nil, nil
	}

	if data[0]|0x20 == 't' {
		data = data[1:]
		p.state = eHTT
		goto protoHTT
	}

	return parser.Error, nil, status.ErrUnsupportedProtocol

protoHTT:
	if len(data) == 0 {
		return parser.Pending, nil, nil
	}

	if data[0]|0x20 == 'p' {
		data = data[1:]
		p.state = eHTTP
		goto protoHTTP
	}

	return parser.Error, nil, status.ErrUnsupportedProtocol

protoHTTP:
	if len(data) == 0 {
		return parser.Pending, nil, nil
	}

	if data[0] == '/' {
		data = data[1:]
		p.state = eProtoMajor
		goto protoMajor
	}

	return parser.Error, nil, status.ErrUnsupportedProtocol

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

	if data[0] == '.' {
		data = data[1:]
		p.state = eProtoMinor
		goto protoMinor
	}

	return parser.Error, nil, status.ErrUnsupportedProtocol

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
		return parser.HeadersCompleted, data[1:], nil
	default:
		// headers are here. I have to have a buffer for header key, and after receiving it,
		// get an appender from headers manager (and keep it in httpRequestsParser struct)
		p.state = eHeaderKey
		goto headerKey
	}

protoCRLFCR:
	if len(data) == 0 {
		return parser.Pending, nil, nil
	}

	if data[0] == '\n' {
		return parser.HeadersCompleted, data[1:], nil
	}

	return parser.Error, nil, status.ErrBadRequest

headerKey:
	if len(data) == 0 {
		return parser.Pending, nil, err
	}

	switch data[0] {
	case '\n':
		return parser.HeadersCompleted, data[1:], nil
	case '\r':
		data = data[1:]
		p.state = eHeaderValueCRLFCR
		goto headerValueCRLFCR
	}

	{
		colon := bytes.IndexByte(data, ':')
		if colon == -1 {
			if !p.headerKeyArena.Append(data...) {
				return parser.Error, nil, status.ErrHeaderFieldsTooLarge
			}

			return parser.Pending, nil, nil
		}

		if !p.headerKeyArena.Append(data[:colon]...) {
			return parser.Error, nil, status.ErrHeaderFieldsTooLarge
		}

		p.headerKey = uf.B2S(p.headerKeyArena.Finish())
		data = data[colon+1:]

		if p.headersNumber++; p.headersNumber > p.headersSettings.Number.Maximal {
			return parser.Error, nil, status.ErrTooManyHeaders
		}

		if strcomp.EqualFold(p.headerKey, "content-length") {
			p.state = eContentLength
			goto contentLength
		}

		p.state = eHeaderValue
		goto headerValue
	}

contentLength:
	for i, char := range data {
		if char == ' ' {
			continue
		}

		if char < '0' || char > '9' {
			data = data[i:]
			goto contentLengthEnd
		}

		p.contentLength = p.contentLength*10 + int(char-'0')
	}

	return parser.Pending, nil, nil

contentLengthEnd:
	// guaranteed, that data at this point contains AT LEAST 1 byte.
	// The proof is, that this code is reachable ONLY if loop has reached a non-digit
	// ascii symbol. In case loop has finished peacefully, as no more data left, but also no
	// character found to satisfy the exit condition, this code will never be reached
	p.request.ContentLength = p.contentLength

	switch data[0] {
	case ' ':
	case '\r':
		data = data[1:]
		p.state = eContentLengthCR
		goto contentLengthCR
	case '\n':
		data = data[1:]
		p.state = eHeaderKey
		goto headerKey
	default:
		return parser.Error, nil, status.ErrBadRequest
	}

contentLengthCR:
	if len(data) == 0 {
		return parser.Pending, nil, nil
	}

	if data[0] != '\n' {
		return parser.Error, nil, status.ErrBadRequest
	}

	data = data[1:]
	p.state = eHeaderKey
	goto headerKey

contentLengthCRLFCR:
	if len(data) == 0 {
		return parser.Pending, nil, nil
	}

	if data[0] == '\n' {
		return parser.HeadersCompleted, data[1:], nil
	}

	return parser.Error, nil, status.ErrBadRequest

headerValue:
	{
		lf := bytes.IndexByte(data, '\n')
		if lf == -1 {
			if !p.headerValueArena.Append(data...) {
				return parser.Error, nil, status.ErrHeaderFieldsTooLarge
			}

			if p.headerValueArena.SegmentLength() > p.headersSettings.MaxValueLength {
				return parser.Error, nil, status.ErrHeaderFieldsTooLarge
			}

			return parser.Pending, nil, nil
		}

		if !p.headerValueArena.Append(data[:lf]...) {
			return parser.Error, nil, status.ErrHeaderFieldsTooLarge
		}

		if p.headerValueArena.SegmentLength() > p.headersSettings.MaxValueLength {
			return parser.Error, nil, status.ErrHeaderFieldsTooLarge
		}

		data = data[lf+1:]
		value := uf.B2S(trimPrefixSpaces(p.headerValueArena.Finish()))
		if value[len(value)-1] == '\r' {
			value = value[:len(value)-1]
		}

		requestHeaders.Add(p.headerKey, value)

		switch {
		case strcomp.EqualFold(p.headerKey, "content-type"):
			p.request.ContentType = value
		case strcomp.EqualFold(p.headerKey, "upgrade"):
			p.request.Upgrade = proto.ChooseUpgrade(value)
		case strcomp.EqualFold(p.headerKey, "transfer-encoding"):
			te := parseTransferEncoding(value)
			te.HasTrailer = p.request.TransferEncoding.HasTrailer
			p.request.TransferEncoding = te
		case strcomp.EqualFold(p.headerKey, "trailer"):
			p.request.TransferEncoding.HasTrailer = true
		}

		p.state = eHeaderKey
		goto headerKey
	}

headerValueCRLFCR:
	if len(data) == 0 {
		return parser.Pending, nil, nil
	}

	if data[0] == '\n' {
		return parser.HeadersCompleted, data[1:], nil
	}

	return parser.Error, nil, status.ErrBadRequest
}

func (p *httpRequestsParser) Release() {
	p.request.Headers.Clear()
	p.protoMajor = 0
	p.protoMinor = 0
	p.headersNumber = 0
	p.begin = 0
	p.pointer = 0
	p.headerKeyArena.Clear()
	p.headerValueArena.Clear()
	p.contentLength = 0
	p.state = eMethod
}

func parseTransferEncoding(value string) (te headers.TransferEncoding) {
	var offset int

	for i := range value {
		if value[i] == ',' {
			te = processTEToken(value[offset:i], te)
			offset = i + 1
		}
	}

	return processTEToken(value[offset:], te)
}

func processTEToken(rawToken string, te headers.TransferEncoding) headers.TransferEncoding {
	switch token := strings.TrimSpace(rawToken); token {
	case "":
	case "chunked":
		te.Chunked = true
	default:
		te.Token = token
	}

	return te
}

func trimPrefixSpaces(b []byte) []byte {
	for i, char := range b {
		if char != ' ' {
			return b[i:]
		}
	}

	return b[:0]
}
