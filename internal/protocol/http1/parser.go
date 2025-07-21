package http1

import (
	"bytes"
	"github.com/flrdv/uf"
	"github.com/indigo-web/indigo/config"
	"github.com/indigo-web/indigo/http"
	"github.com/indigo-web/indigo/http/method"
	"github.com/indigo-web/indigo/http/proto"
	"github.com/indigo-web/indigo/http/status"
	"github.com/indigo-web/indigo/internal/buffer"
	"github.com/indigo-web/indigo/internal/hexconv"
	"strings"
)

type parserState uint8

const (
	eMethod parserState = iota + 1
	ePath
	ePathDecode1Char
	ePathDecode2Char
	eParamsKey
	eParamsKeyDecode1Char
	eParamsKeyDecode2Char
	eParamsValue
	eParamsValueDecode1Char
	eParamsValueDecode2Char
	eProtocol
	eHeaderKey
	eContentLength
	eContentLengthCR
	eHeaderValue
	eHeaderValueCRLFCR
)

type parser struct {
	urlEncodedChar    uint8
	state             parserState
	cfg               *config.Config
	request           *http.Request
	headersNumber     int
	contentLength     int
	key               string
	transferEncodings []string
	contentEncodings  []string
	requestLine       *buffer.Buffer
	headerKeys        *buffer.Buffer
	headerValues      *buffer.Buffer
}

func newParser(
	cfg *config.Config, request *http.Request, headerKeys, headerValues, requestLine buffer.Buffer,
) *parser {
	return &parser{
		cfg:     cfg,
		state:   eMethod,
		request: request,
		// TODO: pass these through arguments instead of allocating in-place
		transferEncodings: make([]string, 0, cfg.Headers.MaxEncodingTokens),
		contentEncodings:  make([]string, 0, cfg.Headers.MaxEncodingTokens),
		requestLine:       &requestLine,
		headerKeys:        &headerKeys,
		headerValues:      &headerValues,
	}
}

func (p *parser) Parse(data []byte) (done bool, extra []byte, err error) {
	_ = *p.request
	request := p.request
	requestLine := p.requestLine
	headerKeys := p.headerKeys
	headerValues := p.headerValues
	headersCfg := p.cfg.Headers

	switch p.state {
	case eMethod:
		goto method
	case ePath:
		goto path
	case ePathDecode1Char:
		goto pathDecode1Char
	case ePathDecode2Char:
		goto pathDecode2Char
	case eParamsKey:
		goto paramsKey
	case eParamsKeyDecode1Char:
		goto paramsKeyDecode1Char
	case eParamsKeyDecode2Char:
		goto paramsKeyDecode2Char
	case eParamsValue:
		goto paramsValue
	case eParamsValueDecode1Char:
		goto paramsValueDecode1Char
	case eParamsValueDecode2Char:
		goto paramsValueDecode2Char
	case eProtocol:
		goto protocol
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
		panic("unreachable code")
	}

method:
	for i := 0; i < len(data); i++ {
		if data[i] == ' ' {
			var methodValue []byte
			if requestLine.SegmentLength() == 0 {
				methodValue = data[:i]
			} else {
				if !requestLine.Append(data[:i]) {
					return true, nil, status.ErrTooLongRequestLine
				}

				methodValue = requestLine.Preview()
				requestLine.Discard(0)
			}

			if len(methodValue) == 0 {
				return true, nil, status.ErrBadRequest
			}

			request.Method = method.Parse(uf.B2S(methodValue))
			if request.Method == method.Unknown {
				return true, nil, status.ErrMethodNotImplemented
			}

			data = data[i+1:]
			goto path
		}
	}

	if !requestLine.Append(data) {
		return true, nil, status.ErrMethodNotImplemented
	}

	return false, nil, nil

path:
	{
		checkpoint := 0

		for i := 0; i < len(data); i++ {
			switch char := data[i]; char {
			case '%':
				if !requestLine.Append(data[checkpoint:i]) {
					return true, nil, status.ErrURITooLong
				}

				if len(data[i+1:]) >= 2 {
					// fast path
					c := (hexconv.Halfbyte[data[i+1]] << 4) | hexconv.Halfbyte[data[i+2]]
					if isProhibitedChar(c) {
						return true, nil, status.ErrBadRequest
					}

					if !requestLine.AppendByte(c) {
						return true, nil, status.ErrURITooLong
					}
				} else {
					// slow path
					goto pathDecode1Char
				}

				i += 2
				checkpoint = i + 1
			case ' ':
				if !requestLine.Append(data[checkpoint:i]) {
					return true, nil, status.ErrURITooLong
				}

				request.Path = uf.B2S(requestLine.Finish())
				if len(request.Path) == 0 {
					return true, nil, status.ErrBadRequest
				}

				data = data[i+1:]
				goto protocol
			case '?':
				if !requestLine.Append(data[checkpoint:i]) {
					return true, nil, status.ErrURITooLong
				}

				request.Path = uf.B2S(requestLine.Finish())
				data = data[i+1:]
				goto paramsKey
			case '#':
				// fragments are generally not allowed in request paths. In order to keep the parser
				// compact and not bloat it with unnecessary states, simply reject such requests.
				return true, nil, status.ErrBadRequest
			default:
				if isProhibitedChar(char) {
					return true, nil, status.ErrBadRequest
				}
			}
		}

		if !requestLine.Append(data[checkpoint:]) {
			return true, nil, status.ErrURITooLong
		}

		p.state = ePath
		return false, nil, nil
	}

pathDecode1Char:
	if len(data) == 0 {
		p.state = ePathDecode1Char
		return false, nil, nil
	}

	p.urlEncodedChar, data = data[0], data[1:]
	// fallthrough to pathDecode2Char

pathDecode2Char:
	{
		if len(data) == 0 {
			p.state = ePathDecode2Char
			return false, nil, nil
		}

		char := (hexconv.Halfbyte[p.urlEncodedChar] << 4) | hexconv.Halfbyte[data[0]]
		if isProhibitedChar(char) {
			return true, nil, status.ErrBadRequest
		}

		if !requestLine.AppendByte(char) {
			return true, nil, status.ErrURITooLong
		}

		data = data[1:]
		goto path
	}

paramsKey:
	for i := 0; i < len(data); i++ {
		switch char := data[i]; char {
		case '+':
			if !requestLine.AppendByte(' ') {
				return true, nil, status.ErrTooLongRequestLine
			}
		case '%':
			if len(data[i+1:]) >= 2 {
				// fast path
				c := (hexconv.Halfbyte[data[i+1]] << 4) | hexconv.Halfbyte[data[i+2]]
				if isProhibitedChar(c) {
					return true, nil, status.ErrBadParams
				}

				if !requestLine.AppendByte(c) {
					return true, nil, status.ErrTooLongRequestLine
				}

				i += 2
			} else {
				// slow path
				data = data[i+1:]
				goto paramsKeyDecode1Char
			}
		case '=':
			p.key = uf.B2S(requestLine.Finish())
			data = data[i+1:]
			goto paramsValue
		case ' ':
			request.Params.Add(uf.B2S(requestLine.Finish()), "")
			data = data[i+1:]
			goto protocol
		default:
			if isProhibitedChar(char) {
				return true, nil, status.ErrBadParams
			}

			if !requestLine.AppendByte(char) {
				return true, nil, status.ErrTooLongRequestLine
			}
		}
	}

paramsKeyDecode1Char:
	if len(data) == 0 {
		p.state = eParamsKeyDecode1Char
		return false, nil, nil
	}

	p.urlEncodedChar = data[0]
	data = data[1:]
	// fallthrough to paramsKeyDecode2Char

paramsKeyDecode2Char:
	{
		if len(data) == 0 {
			p.state = eParamsKeyDecode2Char
			return false, nil, nil
		}

		char := (hexconv.Halfbyte[p.urlEncodedChar] << 4) | hexconv.Halfbyte[data[0]]
		if isProhibitedChar(char) {
			return true, nil, status.ErrBadParams
		}

		if !requestLine.AppendByte(char) {
			return true, nil, status.ErrTooLongRequestLine
		}

		goto paramsKey
	}

paramsValue:
	{
		checkpoint := 0

		for i := 0; i < len(data); i++ {
			switch char := data[i]; char {
			case '+':
				// will be flushed later anyway
				data[i] = ' '
			case '%':
				if !requestLine.Append(data[checkpoint:i]) {
					return true, nil, status.ErrTooLongRequestLine
				}

				if len(data[i+1:]) >= 2 {
					// fast path
					c := (hexconv.Halfbyte[data[i+1]] << 4) | hexconv.Halfbyte[data[i+2]]
					if isProhibitedChar(c) {
						return true, nil, status.ErrBadParams
					}

					if !requestLine.AppendByte(c) {
						return true, nil, status.ErrTooLongRequestLine
					}

					i += 2
					checkpoint = i + 1
				} else {
					// slow path
					data = data[i+1:]
					goto paramsValueDecode1Char
				}
			case '&':
				if !requestLine.Append(data[checkpoint:i]) {
					return true, nil, status.ErrTooLongRequestLine
				}

				request.Params.Add(p.key, uf.B2S(requestLine.Finish()))
				data = data[i+1:]
				goto paramsKey
			case ' ':
				if !requestLine.Append(data[checkpoint:i]) {
					return true, nil, status.ErrTooLongRequestLine
				}

				request.Params.Add(p.key, uf.B2S(requestLine.Finish()))
				data = data[i+1:]
				goto protocol
			case '#':
				return true, nil, status.ErrBadRequest
			}
		}

		if !requestLine.Append(data[checkpoint:]) {
			return true, nil, status.ErrTooLongRequestLine
		}

		p.state = eParamsValue
		return false, nil, nil
	}

paramsValueDecode1Char:
	if len(data) == 0 {
		p.state = eParamsValueDecode1Char
		return false, nil, nil
	}

	p.urlEncodedChar = data[0]
	data = data[1:]
	// fallthrough to paramsValueDecode2Char

paramsValueDecode2Char:
	{
		if len(data) == 0 {
			p.state = eParamsValueDecode2Char
			return false, nil, nil
		}

		char := (hexconv.Halfbyte[p.urlEncodedChar] << 4) | hexconv.Halfbyte[data[0]]
		if isProhibitedChar(char) {
			return true, nil, status.ErrBadParams
		}

		if !requestLine.AppendByte(char) {
			return true, nil, status.ErrTooLongRequestLine
		}

		goto paramsValue
	}

protocol:
	{
		boundary := bytes.IndexByte(data, '\n')
		if boundary == -1 {
			if !requestLine.Append(data) {
				return true, nil, status.ErrTooLongRequestLine
			}

			p.state = eProtocol
			return false, nil, nil
		}

		var protocol proto.Protocol
		if requestLine.SegmentLength() == 0 {
			protocol = proto.FromBytes(stripCR(data[:boundary]))
		} else {
			if !requestLine.Append(data[:boundary]) {
				return true, nil, status.ErrTooLongRequestLine
			}

			protocol = proto.FromBytes(stripCR(requestLine.Preview()))
		}

		if protocol == proto.Unknown {
			return true, nil, status.ErrHTTPVersionNotSupported
		}

		request.Protocol = protocol
		data = data[boundary+1:]
		// fallthrough to headerKey
	}

headerKey:
	{
		if len(data) == 0 {
			p.state = eHeaderKey
			return false, nil, nil
		}

		switch data[0] {
		case '\n':
			p.cleanup()

			return true, data[1:], nil
		case '\r':
			data = data[1:]
			goto headerValueCRLFCR
		}

		colon := bytes.IndexByte(data, ':')
		if colon == -1 {
			if !headerKeys.Append(data) {
				return true, nil, status.ErrHeaderFieldsTooLarge
			}

			p.state = eHeaderKey
			return false, nil, nil
		}

		if !headerKeys.Append(data[:colon]) {
			return true, nil, status.ErrHeaderFieldsTooLarge
		}

		key := uf.B2S(headerKeys.Finish())
		p.key = key
		data = data[colon+1:]

		if p.headersNumber++; p.headersNumber > headersCfg.Number.Maximal {
			return true, nil, status.ErrTooManyHeaders
		}

		if len(key) == len("content-length") &&
			cContent == encodeU64(
				key[0]|0x20, key[1]|0x20, key[2]|0x20, key[3]|0x20, key[4]|0x20, key[5]|0x20, key[6]|0x20, key[7]|0x20,
			) && cLength == encodeU64(
			key[8]|0x20, key[9]|0x20, key[10]|0x20, key[11]|0x20, key[12]|0x20, key[13]|0x20, 0, 0,
		) {
			goto contentLength
		}

		// fallthrough to headerValue
	}

headerValue:
	{
		lf := bytes.IndexByte(data, '\n')
		if lf == -1 {
			if !headerValues.Append(data) {
				return true, nil, status.ErrHeaderFieldsTooLarge
			}

			if headerValues.SegmentLength() > headersCfg.MaxValueLength {
				return true, nil, status.ErrHeaderFieldsTooLarge
			}

			p.state = eHeaderValue
			return false, nil, nil
		}

		if !headerValues.Append(data[:lf]) {
			return true, nil, status.ErrHeaderFieldsTooLarge
		}

		if headerValues.Preview()[headerValues.SegmentLength()-1] == '\r' {
			headerValues.Trunc(1)
		}

		if headerValues.SegmentLength() > headersCfg.MaxValueLength {
			return true, nil, status.ErrHeaderFieldsTooLarge
		}

		data = data[lf+1:]
		value := uf.B2S(trimPrefixSpaces(headerValues.Finish()))

		key := p.key
		request.Headers.Add(key, value)

		switch len(key) {
		case 7:
			if cUpgrade == encodeU64(
				key[0]|0x20, key[1]|0x20, key[2]|0x20, key[3]|0x20, key[4]|0x20, key[5]|0x20, key[6]|0x20, 0,
			) {
				request.Upgrade = proto.ChooseUpgrade(value)
			}
		case 10:
			if cConnecti == encodeU64(
				key[0]|0x20, key[1]|0x20, key[2]|0x20, key[3]|0x20, key[4]|0x20, key[5]|0x20, key[6]|0x20, key[7]|0x20,
			) && cOn == encodeU16(key[8]|0x20, key[9]|0x20) {
				request.Connection = value
			}
		case 12:
			if cContent == encodeU64(
				key[0]|0x20, key[1]|0x20, key[2]|0x20, key[3]|0x20, key[4]|0x20, key[5]|0x20, key[6]|0x20, key[7]|0x20,
			) && cType == encodeU32(
				key[8]|0x20, key[9]|0x20, key[10]|0x20, key[11]|0x20,
			) {
				request.ContentType = value
			}
		case 16:
			if cContent == encodeU64(
				key[0]|0x20, key[1]|0x20, key[2]|0x20, key[3]|0x20, key[4]|0x20, key[5]|0x20, key[6]|0x20, key[7]|0x20,
			) && cEncoding == encodeU64(
				key[8]|0x20, key[9]|0x20, key[10]|0x20, key[11]|0x20, key[12]|0x20, key[13]|0x20, key[14]|0x20, key[15]|0x20,
			) {
				request.Encoding.Content, err = parseEncodingString(p.contentEncodings, value)
				if err != nil {
					return true, nil, err
				}
			}
		case 17:
			if cTransfer == encodeU64(
				key[0]|0x20, key[1]|0x20, key[2]|0x20, key[3]|0x20, key[4]|0x20, key[5]|0x20, key[6]|0x20, key[7]|0x20,
			) && cEncodin == encodeU64(
				key[8]|0x20, key[9]|0x20, key[10]|0x20, key[11]|0x20, key[12]|0x20, key[13]|0x20, key[14]|0x20, key[15]|0x20,
			) && key[16]|0x20 == 'g' {
				request.Encoding.Transfer, err = parseEncodingString(p.transferEncodings, value)
				if err != nil {
					return true, nil, err
				}

				te := request.Encoding.Transfer
				if len(te) > 0 {
					if te[len(te)-1] != "chunked" {
						return true, nil, status.ErrBadEncoding
					}
					
					request.Encoding.Chunked = true
				}

			}
		}

		goto headerKey
	}

headerValueCRLFCR:
	if len(data) == 0 {
		p.state = eHeaderValueCRLFCR
		return false, nil, nil
	}

	if data[0] == '\n' {
		p.cleanup()

		return true, data[1:], nil
	}

	return true, nil, status.ErrBadRequest

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

	p.state = eContentLength
	return false, nil, nil

contentLengthEnd:
	// guaranteed, that data at this point contains AT LEAST 1 byte.
	// The proof is, that this code is reachable ONLY if loop has reached a non-digit
	// ascii symbol. In case loop has finished peacefully, as no more data left, but also no
	// character found to satisfy the exit condition, this code will never be reached
	request.ContentLength = p.contentLength

	switch data[0] {
	case '\r':
		data = data[1:]
		goto contentLengthCR
	case '\n':
		data = data[1:]
		goto headerKey
	default:
		return true, nil, status.ErrBadRequest
	}

contentLengthCR:
	if len(data) == 0 {
		p.state = eContentLengthCR
		return false, nil, nil
	}

	if data[0] != '\n' {
		return true, nil, status.ErrBadRequest
	}

	data = data[1:]
	goto headerKey
}

func (p *parser) cleanup() {
	p.headersNumber = 0
	p.requestLine.Clear()
	p.headerKeys.Clear()
	p.headerValues.Clear()
	p.contentLength = 0
	p.transferEncodings = p.transferEncodings[:0]
	p.contentEncodings = p.contentEncodings[:0]
	p.state = eMethod
}

func parseEncodingString(buff []string, value string) (toks []string, err error) {
	var token string

	for len(value) > 0 {
		comma := strings.IndexByte(value, ',')
		if comma == -1 {
			token, value = value, ""
		} else {
			token, value = value[:comma], value[comma+1:]
		}

		token = trimSpaces(token)
		if len(token) == 0 {
			return nil, status.ErrUnsupportedEncoding
		}

		if len(buff) >= cap(buff) {
			return nil, status.ErrTooManyEncodingTokens
		}

		buff = append(buff, token)
	}

	return buff, nil
}

func trimSpaces(s string) string {
	for i, char := range s {
		if char != ' ' {
			s = s[i:]
			break
		}
	}

	for i := len(s); i > 0; i-- {
		if s[i-1] != ' ' {
			return s[:i]
		}
	}

	return s[:0]
}

func trimPrefixSpaces(b []byte) []byte {
	for i, char := range b {
		if char != ' ' {
			return b[i:]
		}
	}

	return b[:0]
}

func stripCR(b []byte) []byte {
	if len(b) > 0 && b[len(b)-1] == '\r' {
		return b[:len(b)-1]
	}

	return b
}

func isProhibitedChar(c byte) bool {
	return c < 0x20 || c > 0x7e
}

var (
	cUpgrade  = encodeU64('u', 'p', 'g', 'r', 'a', 'd', 'e', 0)
	cContent  = encodeU64('c', 'o', 'n', 't', 'e', 'n', 't', '-')
	cConnecti = encodeU64('c', 'o', 'n', 'n', 'e', 'c', 't', 'i')
	cOn       = encodeU16('o', 'n')
	cLength   = encodeU64('l', 'e', 'n', 'g', 't', 'h', 0, 0)
	cType     = encodeU32('t', 'y', 'p', 'e')
	cEncoding = encodeU64('e', 'n', 'c', 'o', 'd', 'i', 'n', 'g')
	cTransfer = encodeU64('t', 'r', 'a', 'n', 's', 'f', 'e', 'r')
	cEncodin  = encodeU64('-', 'e', 'n', 'c', 'o', 'd', 'i', 'n')
)

func encodeU64(a, b, c, d, e, f, g, h uint8) uint64 {
	return (uint64(h) << 56) | (uint64(g) << 48) | (uint64(f) << 40) | (uint64(e) << 32) |
		(uint64(d) << 24) | (uint64(c) << 16) | (uint64(b) << 8) | uint64(a)
}

func encodeU32(a, b, c, d uint8) uint32 {
	return (uint32(d) << 24) | (uint32(c) << 16) | (uint32(b) << 8) | uint32(a)
}

func encodeU16(a, b uint8) uint16 {
	return (uint16(b) << 8) | uint16(a)
}
