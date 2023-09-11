package http1

import (
	"bytes"
	"github.com/indigo-web/indigo/client"
	"github.com/indigo-web/indigo/client/internal/parser"
	"github.com/indigo-web/indigo/http"
	"github.com/indigo-web/indigo/http/headers"
	"github.com/indigo-web/indigo/http/proto"
	"github.com/indigo-web/indigo/http/status"
	"github.com/indigo-web/utils/buffer"
	"github.com/indigo-web/utils/uf"
)

var _ parser.Parser = &Parser{}

type Parser struct {
	state        parserState
	response     client.Response
	respLineBuff buffer.Buffer[byte]
	headersBuff  buffer.Buffer[byte]
	headerKey    string
}

func NewParser(respLineBuff, headersBuff buffer.Buffer[byte]) *Parser {
	return &Parser{
		state:        eProto,
		respLineBuff: respLineBuff,
		headersBuff:  headersBuff,
	}
}

func (p *Parser) Init(headers *headers.Headers, body *http.Body) {
	p.response = client.NewResponse(headers, body)
}

func (p *Parser) Parse(data []byte) (headersCompleted bool, rest []byte, err error) {
	switch p.state {
	case eProto:
		goto proto
	case eCode:
		goto code
	case eStatus:
		goto status
	case eHeaderKey:
		goto headerKey
	case eHeaderKeyCR:
		goto headerKeyCR
	case eHeaderSemicolon:
		goto headerSemicolon
	case eHeaderValue:
		goto headerValue
	default:
		panic("BUG: response parser: unknown state")
	}

proto:
	{
		sp := bytes.IndexByte(data, ' ')
		if sp == -1 {
			if !p.respLineBuff.Append(data...) {
				return false, nil, status.ErrTooLongResponseLine
			}

			return false, nil, nil
		}

		// TODO: if we received the whole protocol all-at-once, we can avoid copying
		//  the data into the buffer and win a bit more of performance
		if !p.respLineBuff.Append(data[:sp]...) {
			return false, nil, status.ErrTooLongResponseLine
		}

		p.response.Protocol = proto.FromBytes(p.respLineBuff.Finish())
		if p.response.Protocol == proto.Unknown {
			return false, nil, status.ErrHTTPVersionNotSupported
		}

		data = data[sp+1:]
		p.state = eCode
		goto code
	}

code:
	for i := 0; i < len(data); i++ {
		if data[i] == ' ' {
			data = data[i+1:]
			p.state = eStatus
			goto status
		}

		if data[i] < '0' || data[i] > '9' {
			return false, nil, status.ErrBadRequest
		}

		p.response.Code = status.Code(int(p.response.Code)*10 + int(data[i]-'0'))
	}

	// note: as status.Code is uint16, and we're not checking overflow, it may
	// actually happen. Other question is, whether it's really anyhow dangerous

	return false, nil, nil

status:
	{
		lf := bytes.IndexByte(data, '\n')
		if lf == -1 {
			if !p.respLineBuff.Append(data...) {
				return false, nil, status.ErrTooLongResponseLine
			}

			return false, nil, nil
		}

		if !p.respLineBuff.Append(data[:lf]...) {
			return false, nil, status.ErrTooLongResponseLine
		}

		p.response.Status = status.Status(uf.B2S(rstripCR(p.respLineBuff.Finish())))
		data = data[lf+1:]
		p.state = eHeaderKey
		goto headerKey
	}

headerKey:
	if len(data) == 0 {
		return false, nil, nil
	}

	switch data[0] {
	case '\r':
		data = data[1:]
		p.state = eHeaderKeyCR
		goto headerKeyCR
	case '\n':
		data = data[1:]
		goto exitSuccess
	}

	{
		semicolon := bytes.IndexByte(data, ':')
		if semicolon == -1 {
			if !p.headersBuff.Append(data...) {
				return false, nil, status.ErrHeaderKeyTooLarge
			}

			return false, nil, nil
		}

		if !p.headersBuff.Append(data[:semicolon]...) {
			return false, nil, status.ErrHeaderKeyTooLarge
		}

		p.headerKey = uf.B2S(p.headersBuff.Finish())
		data = data[semicolon+1:]
		p.state = eHeaderSemicolon
		goto headerSemicolon
	}

headerKeyCR:
	if data[0] != '\n' {
		return true, nil, status.ErrBadRequest
	}

	data = data[1:]
	goto exitSuccess

headerSemicolon:
	for i := 0; i < len(data); i++ {
		if data[i] != ' ' {
			data = data[i:]
			p.state = eHeaderValue
			goto headerValue
		}
	}

	return false, nil, nil

headerValue:
	{
		lf := bytes.IndexByte(data, '\n')
		if lf == -1 {
			if !p.headersBuff.Append(data...) {
				return false, nil, status.ErrHeaderValueTooLarge
			}

			return false, nil, nil
		}

		if !p.headersBuff.Append(data[:lf]...) {
			return false, nil, status.ErrHeaderValueTooLarge
		}

		switch {
		case 
		}

		p.response.Headers.Add(p.headerKey, uf.B2S(rstripCR(p.headersBuff.Finish())))
		data = data[lf+1:]
		p.state = eHeaderKey
		goto headerKey
	}

exitSuccess:
	p.release()

	return true, data, nil
}

func (p *Parser) Response() client.Response {
	return p.response
}

func (p *Parser) release() {
	p.state = eProto
	p.respLineBuff.Clear()
	p.headersBuff.Clear()
}

func rstripCR(b []byte) []byte {
	if b[len(b)-1] == '\r' {
		b = b[:len(b)-1]
	}

	return b
}
