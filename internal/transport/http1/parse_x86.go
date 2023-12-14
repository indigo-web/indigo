//go:build 386 || amd64

package http1

import (
	"bytes"
	"fmt"
	"github.com/indigo-web/indigo/http/method"
	"github.com/indigo-web/indigo/http/proto"
	"github.com/indigo-web/indigo/http/status"
	"github.com/indigo-web/indigo/internal/transport"
	"github.com/indigo-web/indigo/internal/uridecode"
	"github.com/indigo-web/utils/strcomp"
	"github.com/indigo-web/utils/uf"
)

func (p *Parser) Parse(data []byte) (state transport.RequestState, extra []byte, err error) {
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

		p.request.Method = method.Parse(uf.B2S(methodValue))

		if p.request.Method == method.Unknown {
			return transport.Error, nil, status.ErrMethodNotImplemented
		}

		data = data[sp+1:]
		p.state = ePath
		goto path
	}

path:
	{
		lf := bytes.IndexByte(data, '\n')
		if lf == -1 {
			if !p.startLineBuff.Append(data) {
				return transport.Error, nil, status.ErrURITooLong
			}

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
			p.request.Query.Set(reqPath[query+1:])
			reqPath = reqPath[:query]
		}

		if len(reqPath) == 0 {
			return transport.Error, nil, status.ErrBadRequest
		}

		reqPath, err = uridecode.Decode(reqPath, reqPath[:0])
		if err != nil {
			return transport.Error, nil, err
		}

		p.request.Path = uf.B2S(reqPath)
		p.request.Proto = proto.FromBytes(reqProto)
		if p.request.Proto == proto.Unknown {
			return transport.Error, nil, status.ErrUnsupportedProtocol
		}

		data = data[lf+1:]
		p.state = eHeaderKey
		goto headerKey
	}

	return transport.Pending, nil, nil

headerKey:
	if len(data) == 0 {
		return transport.Pending, nil, err
	}

	switch data[0] {
	case '\n':
		p.reset()

		return transport.HeadersCompleted, data[1:], nil
	case '\r':
		data = data[1:]
		p.state = eHeaderValueCRLFCR
		goto headerValueCRLFCR
	}

	{
		colon := bytes.IndexByte(data, ':')
		if colon == -1 {
			if !p.headerKeyBuff.Append(data) {
				return transport.Error, nil, status.ErrHeaderFieldsTooLarge
			}

			return transport.Pending, nil, nil
		}

		if !p.headerKeyBuff.Append(data[:colon]) {
			return transport.Error, nil, status.ErrHeaderFieldsTooLarge
		}

		p.headerKey = uf.B2S(p.headerKeyBuff.Finish())
		data = data[colon+1:]

		if p.headersNumber++; p.headersNumber > p.headersSettings.Number.Maximal {
			return transport.Error, nil, status.ErrTooManyHeaders
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

	return transport.Pending, nil, nil

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
		return transport.Error, nil, status.ErrBadRequest
	}

contentLengthCR:
	if len(data) == 0 {
		return transport.Pending, nil, nil
	}

	if data[0] != '\n' {
		return transport.Error, nil, status.ErrBadRequest
	}

	data = data[1:]
	p.state = eHeaderKey
	goto headerKey

headerValue:
	{
		lf := bytes.IndexByte(data, '\n')
		if lf == -1 {
			if !p.headerValueBuff.Append(data) {
				return transport.Error, nil, status.ErrHeaderFieldsTooLarge
			}

			if p.headerValueBuff.SegmentLength() > p.headersSettings.MaxValueLength {
				return transport.Error, nil, status.ErrHeaderFieldsTooLarge
			}

			return transport.Pending, nil, nil
		}

		if !p.headerValueBuff.Append(data[:lf]) {
			return transport.Error, nil, status.ErrHeaderFieldsTooLarge
		}

		if p.headerValueBuff.SegmentLength() > p.headersSettings.MaxValueLength {
			return transport.Error, nil, status.ErrHeaderFieldsTooLarge
		}

		data = data[lf+1:]
		value := uf.B2S(trimPrefixSpaces(p.headerValueBuff.Finish()))
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
			p.request.Encoding.Transfer = parseEncodingString(p.encToksBuff, value, cap(p.encToksBuff))
		case strcomp.EqualFold(p.headerKey, "content-encoding"):
			p.request.Encoding.Content = parseEncodingString(p.encToksBuff, value, cap(p.encToksBuff))
		case strcomp.EqualFold(p.headerKey, "trailer"):
			p.request.Encoding.HasTrailer = true
		}

		p.state = eHeaderKey
		goto headerKey
	}

headerValueCRLFCR:
	if len(data) == 0 {
		return transport.Pending, nil, nil
	}

	if data[0] == '\n' {
		p.reset()

		return transport.HeadersCompleted, data[1:], nil
	}

	return transport.Error, nil, status.ErrBadRequest
}
