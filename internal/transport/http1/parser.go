package http1

import (
	"bytes"
	"fmt"
	"github.com/indigo-web/indigo/http"
	"github.com/indigo-web/indigo/http/method"
	"github.com/indigo-web/indigo/http/proto"
	"github.com/indigo-web/indigo/http/status"
	"github.com/indigo-web/indigo/internal/transport"
	"github.com/indigo-web/indigo/internal/uridecode"
	"github.com/indigo-web/indigo/settings"
	"github.com/indigo-web/utils/buffer"
	"github.com/indigo-web/utils/strcomp"
	"github.com/indigo-web/utils/uf"
	"strings"
)

type parserState uint8

const (
	eMethod parserState = iota + 1
	ePath
	eHeaderKey
	eContentLength
	eContentLengthCR
	eHeaderValue
	eHeaderValueCRLFCR
)

// Parser is a stream-based http requests transport. It modifies
// request object by pointer in performance purposes. Decodes query-encoded
// values by its own, you can see that by presented states ePathDecode1Char,
// ePathDecode2Char, etc. When headers are parsed, parser returns state
// transport.HeadersCompleted to notify http server about this, attaching all
// the pending data as an extra. Body must be processed separately
type Parser struct {
	request         *http.Request
	startLineBuff   *buffer.Buffer
	headerKeyBuff   *buffer.Buffer
	headerValueBuff *buffer.Buffer
	encToksBuff     []string
	contEncToksBuff []string
	headerKey       string
	headersSettings *settings.Headers
	headersNumber   int
	contentLength   int
	urlEncodedChar  uint8
	state           parserState
}

func NewParser(
	request *http.Request, keyBuff, valBuff, startLineBuff *buffer.Buffer, headersSettings settings.Headers,
) *Parser {
	return &Parser{
		state:           eMethod,
		request:         request,
		headersSettings: &headersSettings,
		startLineBuff:   startLineBuff,
		encToksBuff:     make([]string, 0, headersSettings.MaxEncodingTokens),
		contEncToksBuff: make([]string, 0, headersSettings.MaxEncodingTokens),
		headerKeyBuff:   keyBuff,
		headerValueBuff: valBuff,
	}
}

func (p *Parser) Parse(data []byte) (state transport.RequestState, extra []byte, err error) {
	_ = *p.request
	request := p.request
	headerKeyBuff := p.headerKeyBuff
	headerValueBuff := p.headerValueBuff

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
			if !p.startLineBuff.Append(data) {
				return transport.Error, nil, status.ErrTooLongRequestLine
			}

			return transport.Pending, nil, nil
		}

		var methodValue []byte
		if p.startLineBuff.SegmentLength() == 0 {
			methodValue = data[:sp]
		} else {
			if !p.startLineBuff.Append(data[:sp]) {
				return transport.Error, nil, status.ErrTooLongRequestLine
			}

			methodValue = p.startLineBuff.Finish()
		}

		if len(methodValue) == 0 {
			return transport.Error, nil, status.ErrBadRequest
		}

		request.Method = method.Parse(uf.B2S(methodValue))
		if request.Method == method.Unknown {
			return transport.Error, nil, status.ErrMethodNotImplemented
		}

		data = data[sp+1:]
		goto path
	}

path:
	{
		lf := bytes.IndexByte(data, '\n')
		if lf == -1 {
			if !p.startLineBuff.Append(data) {
				return transport.Error, nil, status.ErrURITooLong
			}

			p.state = ePath
			return transport.Pending, nil, nil
		}

		if !p.startLineBuff.Append(data[:lf]) {
			return transport.Error, nil, status.ErrURITooLong
		}

		pathAndProto := p.startLineBuff.Finish()
		sp := bytes.LastIndexByte(pathAndProto, ' ')
		if sp == -1 {
			return transport.Error, nil, status.ErrBadRequest
		}

		reqPath, reqProto := pathAndProto[:sp], pathAndProto[sp+1:]
		if reqProto[len(reqProto)-1] == '\r' {
			reqProto = reqProto[:len(reqProto)-1]
		}

		query := bytes.IndexByte(reqPath, '?')
		if query != -1 {
			request.Query.Set(reqPath[query+1:])
			reqPath = reqPath[:query]
		}

		if len(reqPath) == 0 {
			return transport.Error, nil, status.ErrBadRequest
		}

		reqPath, err = uridecode.Decode(reqPath, reqPath[:0])
		if err != nil {
			return transport.Error, nil, err
		}

		request.Path = uf.B2S(reqPath)
		request.Proto = proto.FromBytes(reqProto)
		if request.Proto == proto.Unknown {
			return transport.Error, nil, status.ErrUnsupportedProtocol
		}

		data = data[lf+1:]
		goto headerKey
	}

headerKey:
	{
		if len(data) == 0 {
			p.state = eHeaderKey
			return transport.Pending, nil, err
		}

		switch data[0] {
		case '\n':
			p.reset()

			return transport.HeadersCompleted, data[1:], nil
		case '\r':
			data = data[1:]
			goto headerValueCRLFCR
		}

		colon := bytes.IndexByte(data, ':')
		if colon == -1 {
			if !headerKeyBuff.Append(data) {
				return transport.Error, nil, status.ErrHeaderFieldsTooLarge
			}

			p.state = eHeaderKey
			return transport.Pending, nil, nil
		}

		if !headerKeyBuff.Append(data[:colon]) {
			return transport.Error, nil, status.ErrHeaderFieldsTooLarge
		}

		key := uf.B2S(headerKeyBuff.Finish())
		p.headerKey = key
		data = data[colon+1:]

		if p.headersNumber++; p.headersNumber > p.headersSettings.Number.Maximal {
			return transport.Error, nil, status.ErrTooManyHeaders
		}

		if len(key) == len("content-length") &&
			cContent == encodeU64(
				key[0]|0x20, key[1]|0x20, key[2]|0x20, key[3]|0x20, key[4]|0x20, key[5]|0x20, key[6]|0x20, key[7]|0x20,
			) && cLength == encodeU64(
			key[8]|0x20, key[9]|0x20, key[10]|0x20, key[11]|0x20, key[12]|0x20, key[13]|0x20, 0, 0,
		) {
			goto contentLength
		}

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

	p.state = eContentLength
	return transport.Pending, nil, nil

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
		return transport.Error, nil, status.ErrBadRequest
	}

contentLengthCR:
	if len(data) == 0 {
		p.state = eContentLengthCR
		return transport.Pending, nil, nil
	}

	if data[0] != '\n' {
		return transport.Error, nil, status.ErrBadRequest
	}

	data = data[1:]
	goto headerKey

headerValue:
	{
		lf := bytes.IndexByte(data, '\n')
		if lf == -1 {
			if !headerValueBuff.Append(data) {
				return transport.Error, nil, status.ErrHeaderFieldsTooLarge
			}

			if headerValueBuff.SegmentLength() > p.headersSettings.MaxValueLength {
				return transport.Error, nil, status.ErrHeaderFieldsTooLarge
			}

			p.state = eHeaderValue
			return transport.Pending, nil, nil
		}

		if !headerValueBuff.Append(data[:lf]) {
			return transport.Error, nil, status.ErrHeaderFieldsTooLarge
		}

		if headerValueBuff.SegmentLength() > p.headersSettings.MaxValueLength {
			return transport.Error, nil, status.ErrHeaderFieldsTooLarge
		}

		data = data[lf+1:]
		value := uf.B2S(trimPrefixSpaces(headerValueBuff.Finish()))
		if value[len(value)-1] == '\r' {
			value = value[:len(value)-1]
		}

		request.Headers.Add(p.headerKey, value)

		switch key := p.headerKey; len(p.headerKey) {
		case 7:
			encoded := encodeU64(
				key[0]|0x20, key[1]|0x20, key[2]|0x20, key[3]|0x20, key[4]|0x20, key[5]|0x20, key[6]|0x20, 0,
			)

			switch encoded {
			case cUpgrade:
				request.Upgrade = proto.ChooseUpgrade(value)
			case cTrailer:
				request.Encoding.HasTrailer = true
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
				request.Encoding.Content, _ = parseEncodingString(p.contEncToksBuff, value, cap(p.contEncToksBuff))
			}
		case 17:
			if cTransfer == encodeU64(
				key[0]|0x20, key[1]|0x20, key[2]|0x20, key[3]|0x20, key[4]|0x20, key[5]|0x20, key[6]|0x20, key[7]|0x20,
			) && cEncodin == encodeU64(
				key[8]|0x20, key[9]|0x20, key[10]|0x20, key[11]|0x20, key[12]|0x20, key[13]|0x20, key[14]|0x20, key[15]|0x20,
			) && key[16]|0x20 == 'g' {
				request.Encoding.Transfer, request.Encoding.Chunked = parseEncodingString(
					p.encToksBuff, value, cap(p.encToksBuff),
				)
			}
		}

		goto headerKey
	}

headerValueCRLFCR:
	if len(data) == 0 {
		p.state = eHeaderValueCRLFCR
		return transport.Pending, nil, nil
	}

	if data[0] == '\n' {
		p.reset()

		return transport.HeadersCompleted, data[1:], nil
	}

	return transport.Error, nil, status.ErrBadRequest
}

func (p *Parser) reset() {
	p.headersNumber = 0
	p.startLineBuff.Clear()
	p.headerKeyBuff.Clear()
	p.headerValueBuff.Clear()
	p.contentLength = 0
	p.encToksBuff = p.encToksBuff[:0]
	p.contEncToksBuff = p.contEncToksBuff[:0]
	p.state = eMethod
}

func parseEncodingString(buff []string, value string, maxTokens int) (toks []string, chunked bool) {
	for len(value) > 0 {
		var token string
		comma := strings.IndexByte(value, ',')
		if comma == -1 {
			token, value = value, ""
		} else {
			token, value = value[:comma], value[comma+1:]
		}

		token = strings.TrimSpace(token)
		if len(token) == 0 {
			continue
		}

		if len(buff)+1 > maxTokens {
			return nil, false
		}

		if strcomp.EqualFold(token, "chunked") {
			chunked = true
		}

		buff = append(buff, token)
	}

	return buff, chunked
}

func trimPrefixSpaces(b []byte) []byte {
	for i, char := range b {
		if char != ' ' {
			return b[i:]
		}
	}

	return b[:0]
}

var (
	cUpgrade  = encodeU64('u', 'p', 'g', 'r', 'a', 'd', 'e', 0)
	cTrailer  = encodeU64('t', 'r', 'a', 'i', 'l', 'e', 'r', 0)
	cContent  = encodeU64('c', 'o', 'n', 't', 'e', 'n', 't', '-')
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
