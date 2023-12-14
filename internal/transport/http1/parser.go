package http1

import (
	"strings"

	"github.com/indigo-web/indigo/http"
	"github.com/indigo-web/indigo/settings"
	"github.com/indigo-web/utils/buffer"
	"github.com/indigo-web/utils/pool"
)

type parserState uint8

const (
	eMethod parserState = iota + 1
	ePath
	ePathDecode1Char
	ePathDecode2Char
	eQuery
	eQueryDecode1Char
	eQueryDecode2Char
	eFragment
	eProto
	eProtoEnd
	eProtoCR
	eProtoCRLFCR
	eHeaderKey
	eContentLength
	eContentLengthCR
	eContentLengthCRLFCR
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
	request           *http.Request
	startLineBuff     buffer.Buffer
	encToksBuff       []string
	headerKey         string
	headersValuesPool pool.ObjectPool[[]string]
	headerKeyBuff     buffer.Buffer
	headerValueBuff   buffer.Buffer
	headersSettings   settings.Headers
	headersNumber     int
	contentLength     int
	urlEncodedChar    uint8
	state             parserState
}

func NewParser(
	request *http.Request, keyBuff, valBuff, startLineBuff buffer.Buffer,
	valuesPool pool.ObjectPool[[]string], headersSettings settings.Headers,
) *Parser {
	return &Parser{
		state:             eMethod,
		request:           request,
		headersSettings:   headersSettings,
		startLineBuff:     startLineBuff,
		encToksBuff:       make([]string, 0, headersSettings.MaxEncodingTokens),
		headerKeyBuff:     keyBuff,
		headerValueBuff:   valBuff,
		headersValuesPool: valuesPool,
	}
}

func (p *Parser) reset() {
	p.headersNumber = 0
	p.headerKeyBuff.Clear()
	p.headerValueBuff.Clear()
	p.startLineBuff.Clear()
	p.contentLength = 0
	p.encToksBuff = p.encToksBuff[:0]
	p.state = eMethod
}

func parseEncodingString(buff []string, value string, maxTokens int) (toks []string) {
	var offset int

	for i := 0; i < len(value); i++ {
		if value[i] == ',' {
			token := strings.TrimSpace(value[offset:i])
			offset = i + 1

			if len(token) == 0 {
				continue
			}

			if len(buff)+1 > maxTokens {
				return nil
			}

			buff = append(buff, token)
		}
	}

	token := strings.TrimSpace(value[offset:])
	if len(token) > 0 {
		buff = append(buff, token)
	}

	return buff
}

func trimPrefixSpaces(b []byte) []byte {
	for i, char := range b {
		if char != ' ' {
			return b[i:]
		}
	}

	return b[:0]
}

func isHex(char byte) bool {
	switch {
	case '0' <= char && char <= '9':
		return true
	case 'a' <= char && char <= 'f':
		return true
	case 'A' <= char && char <= 'F':
		return true
	}
	return false
}

func unHex(char byte) byte {
	switch {
	case '0' <= char && char <= '9':
		return char - '0'
	case 'a' <= char && char <= 'f':
		return char - 'a' + 10
	case 'A' <= char && char <= 'F':
		return char - 'A' + 10
	}
	return 0
}
