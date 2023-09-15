package http1

import (
	"bytes"
	"fmt"
	"github.com/indigo-web/indigo/internal/strcomp"
	"github.com/indigo-web/indigo/internal/uridecode"
	"strings"

	"github.com/indigo-web/indigo/http"
	"github.com/indigo-web/indigo/http/method"
	"github.com/indigo-web/indigo/http/proto"
	"github.com/indigo-web/indigo/http/status"
	"github.com/indigo-web/indigo/internal/parser"
	"github.com/indigo-web/indigo/settings"
	"github.com/indigo-web/utils/buffer"
	"github.com/indigo-web/utils/pool"
	"github.com/indigo-web/utils/uf"
)

// httpRequestsParser is a stream-based http requests parser. It modifies
// request object by pointer in performance purposes. Decodes query-encoded
// values by its own, you can see that by presented states ePathDecode1Char,
// ePathDecode2Char, etc. When headers are parsed, parser returns state
// parser.HeadersCompleted to notify http server about this, attaching all
// the pending data as an extra. Body must be processed separately
type httpRequestsParser struct {
	request           *http.Request
	startLineArena    buffer.Buffer[byte]
	encToksBuff       []string
	headerKey         string
	headersValuesPool pool.ObjectPool[[]string]
	headerKeyArena    buffer.Buffer[byte]
	headerValueArena  buffer.Buffer[byte]
	headersSettings   settings.Headers
	headersNumber     int
	contentLength     int
	urlEncodedChar    uint8
	state             parserState
}

func NewHTTPRequestsParser(
	request *http.Request, keyArena, valArena, startLineArena buffer.Buffer[byte],
	valuesPool pool.ObjectPool[[]string], headersSettings settings.Headers,
) parser.HTTPRequestsParser {
	return &httpRequestsParser{
		state:             eMethod,
		request:           request,
		headersSettings:   headersSettings,
		startLineArena:    startLineArena,
		encToksBuff:       make([]string, 0, headersSettings.MaxEncodingTokens),
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
	case eHeaderKey:
		goto headerKey
	case eContentLength:
		goto contentLength
	case eContentLengthCR:
		goto contentLengthCR
	case eHeaderValue:
		goto headerValue
	case eHeaderValueCRLFCR:
		goto headerValueCRLFCR
	default:
		panic(fmt.Sprintf("BUG: unexpected state: %v", p.state))
	}

method:
	{
		sp := bytes.IndexByte(data, ' ')
		if sp == -1 {
			if !p.startLineArena.Append(data...) {
				return parser.Error, nil, status.ErrTooLongRequestLine
			}

			return parser.Pending, nil, nil
		}

		var methodValue []byte
		if p.startLineArena.SegmentLength() == 0 {
			methodValue = data[:sp]
		} else {
			if !p.startLineArena.Append(data[:sp]...) {
				return parser.Error, nil, status.ErrTooLongRequestLine
			}

			methodValue = p.startLineArena.Finish()
		}

		if len(methodValue) == 0 {
			return parser.Error, nil, status.ErrBadRequest
		}

		p.request.Method = method.Parse(uf.B2S(methodValue))

		if p.request.Method == method.Unknown {
			return parser.Error, nil, status.ErrMethodNotImplemented
		}

		data = data[sp+1:]
		p.state = ePath
		goto path
	}

path:
	{
		lf := bytes.IndexByte(data, '\n')
		if lf == -1 {
			if !p.startLineArena.Append(data...) {
				return parser.Error, nil, status.ErrURITooLong
			}

			return parser.Pending, nil, nil
		}

		if !p.startLineArena.Append(data[:lf]...) {
			return parser.Error, nil, status.ErrURITooLong
		}

		pathAndProto := p.startLineArena.Finish()
		sp := bytes.LastIndexByte(pathAndProto, ' ')
		if sp == -1 {
			return parser.Error, nil, status.ErrBadRequest
		}

		reqPath, reqProto := pathAndProto[:sp], pathAndProto[sp+1:]
		if reqProto[len(reqProto)-1] == '\r' {
			reqProto = reqProto[:len(reqProto)-1]
		}

		query := bytes.IndexByte(reqPath, '?')
		if query != -1 {
			p.request.Query.Set(reqPath[query+1:])
			reqPath = reqPath[:query]
		}

		if len(reqPath) == 0 {
			return parser.Error, nil, status.ErrBadRequest
		}

		reqPath, err = uridecode.Decode(reqPath, reqPath[:0])
		if err != nil {
			return parser.Error, nil, err
		}

		p.request.Path = uf.B2S(reqPath)
		p.request.Proto = proto.FromBytes(reqProto)
		if p.request.Proto == proto.Unknown {
			return parser.Error, nil, status.ErrUnsupportedProtocol
		}

		data = data[lf+1:]
		p.state = eHeaderKey
		goto headerKey
	}

	return parser.Pending, nil, nil

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
			tokens, err := p.fillEncoding(value)
			if err != nil {
				return parser.Error, nil, err
			}

			p.request.Encoding.Transfer = tokens
		case strcomp.EqualFold(p.headerKey, "content-encoding"):
			tokens, err := p.fillEncoding(value)
			if err != nil {
				return parser.Error, nil, err
			}

			p.request.Encoding.Content = tokens
		case strcomp.EqualFold(p.headerKey, "trailer"):
			p.request.Encoding.HasTrailer = true
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

func (p *httpRequestsParser) fillEncoding(value string) (tokens []string, err error) {
	tokens, chunked, err := parseEncodingString(p.encToksBuff[:0], value)
	p.request.Encoding.Chunked = p.request.Encoding.Chunked || chunked

	return tokens, err
}

func (p *httpRequestsParser) Release() {
	p.request.Headers.Clear()
	p.headersNumber = 0
	p.headerKeyArena.Clear()
	p.headerValueArena.Clear()
	p.startLineArena.Clear()
	p.contentLength = 0
	p.encToksBuff = p.encToksBuff[:0]
	p.state = eMethod
}

func parseEncodingString(buff []string, value string) (toks []string, chunked bool, err error) {
	var offset int

	for i := range value {
		if value[i] == ',' {
			var isChunked bool
			buff, isChunked, err = processEncodingToken(buff, value[offset:i])
			if err != nil {
				return nil, false, err
			}

			chunked = chunked || isChunked
			offset = i + 1
		}
	}

	var isChunked bool
	buff, isChunked, err = processEncodingToken(buff, value[offset:])

	return buff, chunked || isChunked, err
}

func processEncodingToken(
	buff []string, rawToken string,
) ([]string, bool, error) {
	switch token := strings.TrimSpace(rawToken); token {
	case "":
	case "chunked":
		return buff, true, nil
	default:
		if len(buff)+1 >= cap(buff) {
			return nil, false, status.ErrUnsupportedEncoding
		}

		buff = append(buff, token)
	}

	return buff, false, nil
}

func trimPrefixSpaces(b []byte) []byte {
	for i, char := range b {
		if char != ' ' {
			return b[i:]
		}
	}

	return b[:0]
}
