//go:build !(386 || amd64) || generic_parser

package http1

import (
	"bytes"
	"fmt"
	"github.com/indigo-web/indigo/http/method"
	"github.com/indigo-web/indigo/http/proto"
	"github.com/indigo-web/indigo/http/status"
	"github.com/indigo-web/indigo/internal/transport"
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
	case eProtoCR:
		goto protoCR
	case eProtoEnd:
		goto protoEnd
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
		case ' ':
			if !p.startLineBuff.Append(data[:i]) {
				return transport.Error, nil, status.ErrBadRequest
			}

			if p.startLineBuff.SegmentLength() == 0 {
				return transport.Error, nil, status.ErrBadRequest
			}

			p.request.Method = method.Parse(uf.B2S(p.startLineBuff.Finish()))
			// no need to store the method after we've parsed it
			p.startLineBuff.Clear()

			if p.request.Method == method.Unknown {
				return transport.Error, nil, status.ErrMethodNotImplemented
			}

			data = data[i+1:]
			p.state = ePath
			goto path
		case '\r', '\n':
			// According to rfc2068, 4.1, leading CR or LF should be ignored. However, SIMD parser
			// can't do that and returns just status.ErrBadRequest. In order to stay a bit more consistent,
			// implementing here the same behaviour

			return transport.Error, nil, status.ErrBadRequest
		}
	}

	if !p.startLineBuff.Append(data) {
		return transport.Error, nil, status.ErrBadRequest
	}

	return transport.Pending, nil, nil

path:
	for i := range data {
		switch data[i] {
		case ' ':
			if !p.startLineBuff.Append(data[:i]) {
				return transport.Error, nil, status.ErrBadRequest
			}

			p.request.Path = uf.B2S(p.startLineBuff.Finish())
			if len(p.request.Path) == 0 {
				return transport.Error, nil, status.ErrBadRequest
			}

			data = data[i+1:]
			p.state = eProto
			goto proto
		case '%':
			if !p.startLineBuff.Append(data[:i]) {
				return transport.Error, nil, status.ErrBadRequest
			}

			data = data[i+1:]
			p.state = ePathDecode1Char
			goto pathDecode1Char
		case '?':
			if !p.startLineBuff.Append(data[:i]) {
				return transport.Error, nil, status.ErrBadRequest
			}

			p.request.Path = uf.B2S(p.startLineBuff.Finish())
			if len(p.request.Path) == 0 {
				p.request.Path = "/"
			}

			data = data[i+1:]
			p.state = eQuery
			goto query
		case '#':
			if !p.startLineBuff.Append(data[:i]) {
				return transport.Error, nil, status.ErrBadRequest
			}

			p.request.Path = uf.B2S(p.startLineBuff.Finish())
			if len(p.request.Path) == 0 {
				p.request.Path = "/"
			}

			data = data[i+1:]
			p.state = eFragment
			goto fragment
		case '\x00', '\n', '\r', '\t', '\b', '\a', '\v', '\f':
			// request path MUST NOT include any non-printable characters
			return transport.Error, nil, status.ErrBadRequest
		}
	}

	if !p.startLineBuff.Append(data) {
		return transport.Error, nil, status.ErrBadRequest
	}

	return transport.Pending, nil, nil

pathDecode1Char:
	if len(data) == 0 {
		return transport.Pending, nil, nil
	}

	if !isHex(data[0]) {
		return transport.Error, nil, status.ErrURIDecoding
	}

	p.urlEncodedChar = unHex(data[0]) << 4
	data = data[1:]
	p.state = ePathDecode2Char
	goto pathDecode2Char

pathDecode2Char:
	{
		if len(data) == 0 {
			return transport.Pending, nil, nil
		}

		if !isHex(data[0]) {
			return transport.Error, nil, status.ErrURIDecoding
		}

		var decodedChar [1]byte
		decodedChar[0] = p.urlEncodedChar | unHex(data[0])
		if !p.startLineBuff.Append(decodedChar[:]) {
			return transport.Error, nil, status.ErrURITooLong
		}

		data = data[1:]
		p.state = ePath
		goto path
	}

query:
	for i := range data {
		switch data[i] {
		case ' ':
			if !p.startLineBuff.Append(data[:i]) {
				return transport.Error, nil, status.ErrBadRequest
			}

			p.request.Query.Set(p.startLineBuff.Finish())
			data = data[i+1:]
			p.state = eProto
			goto proto
		case '#':
			if !p.startLineBuff.Append(data[:i]) {
				return transport.Error, nil, status.ErrBadRequest
			}

			p.request.Query.Set(p.startLineBuff.Finish())
			data = data[i+1:]
			p.state = eFragment
			goto fragment
		case '%':
			if !p.startLineBuff.Append(data[:i]) {
				return transport.Error, nil, status.ErrBadRequest
			}

			data = data[i+1:]
			p.state = eQueryDecode1Char
			goto queryDecode1Char
		case '+':
			data[i] = ' '
		case '\x00', '\n', '\r', '\t', '\b', '\a', '\v', '\f':
			return transport.Error, nil, status.ErrBadRequest
		}
	}

	if !p.startLineBuff.Append(data) {
		return transport.Error, nil, status.ErrBadRequest
	}

	return transport.Pending, nil, nil

queryDecode1Char:
	if len(data) == 0 {
		return transport.Pending, nil, nil
	}

	if !isHex(data[0]) {
		return transport.Error, nil, status.ErrURIDecoding
	}

	p.urlEncodedChar = unHex(data[0]) << 4
	data = data[1:]
	p.state = eQueryDecode2Char
	goto queryDecode2Char

queryDecode2Char:
	{
		if len(data) == 0 {
			return transport.Pending, nil, nil
		}

		if !isHex(data[0]) {
			return transport.Error, nil, status.ErrURIDecoding
		}

		var decodedChar [1]byte
		decodedChar[0] = p.urlEncodedChar | unHex(data[0])
		if !p.startLineBuff.Append(decodedChar[:]) {
			return transport.Error, nil, status.ErrURITooLong
		}

		data = data[1:]
		p.state = eQuery
		goto query
	}

fragment:
	{
		sp := bytes.IndexByte(data, ' ')
		if sp == -1 {
			return transport.Pending, nil, nil
		}

		data = data[sp+1:]
		p.state = eProto
		goto proto
	}

proto:
	for i := range data {
		switch data[i] {
		case '\r':
			if !p.startLineBuff.Append(data[:i]) {
				return transport.Error, nil, status.ErrUnsupportedProtocol
			}

			p.request.Proto = proto.FromBytes(p.startLineBuff.Finish())
			if p.request.Proto == proto.Unknown {
				return transport.Error, nil, status.ErrUnsupportedProtocol
			}

			data = data[i+1:]
			p.state = eProtoCR
			goto protoCR
		case '\n':
			if !p.startLineBuff.Append(data[:i]) {
				return transport.Error, nil, status.ErrUnsupportedProtocol
			}

			p.request.Proto = proto.FromBytes(p.startLineBuff.Finish())
			if p.request.Proto == proto.Unknown {
				return transport.Error, nil, status.ErrUnsupportedProtocol
			}

			data = data[i+1:]
			p.state = eProtoEnd
			goto protoEnd
		}
	}

	if !p.startLineBuff.Append(data) {
		return transport.Error, nil, status.ErrUnsupportedProtocol
	}

	return transport.Pending, nil, nil

protoCR:
	if len(data) == 0 {
		return transport.Pending, nil, nil
	}

	if data[0] != '\n' {
		return transport.Error, nil, status.ErrBadRequest
	}

	data = data[1:]
	p.state = eProtoEnd
	goto protoEnd

protoEnd:
	if len(data) == 0 {
		return transport.Pending, nil, nil
	}

	switch data[0] {
	case '\r':
		data = data[1:]
		p.state = eProtoCRLFCR
		goto protoCRLFCR
	case '\n':
		p.reset()

		return transport.HeadersCompleted, data[1:], nil
	}

	p.state = eHeaderKey
	goto headerKey

protoCRLFCR:
	if len(data) == 0 {
		return transport.Pending, nil, nil
	}

	if data[0] == '\n' {
		p.reset()

		return transport.HeadersCompleted, data[1:], nil
	}

	return transport.Error, nil, status.ErrBadRequest

headerKey:
	{
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

contentLengthCRLFCR:
	if len(data) == 0 {
		return transport.Pending, nil, nil
	}

	if data[0] == '\n' {
		p.reset()

		return transport.HeadersCompleted, data[1:], nil
	}

	return transport.Error, nil, status.ErrBadRequest

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
