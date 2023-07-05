package http1

import (
	"fmt"
	"io"

	"github.com/indigo-web/indigo/http/status"
	"github.com/indigo-web/indigo/settings"
)

type ChunkedBodyParser interface {
	Parse(data []byte, trailer bool) (chunk, extra []byte, err error)
}

// chunkedBodyParser is a parser for chunked encoded request bodies
// used to encapsulate process of parsing because it's more convenient
// to leave the process here and let main parser parse only http requests
type chunkedBodyParser struct {
	state chunkedBodyParserState

	settings    settings.Body
	chunkLength int64
}

func NewChunkedBodyParser(settings settings.Body) ChunkedBodyParser {
	parser := newChunkedBodyParser(settings)

	return &parser
}

func newChunkedBodyParser(settings settings.Body) chunkedBodyParser {
	return chunkedBodyParser{
		state:    eChunkLength1Char,
		settings: settings,
	}
}

// Parse a stream of chunked body parts. When fully parsed, nil-chunk is returned, but non-nil
// extra and io.EOF error
func (c *chunkedBodyParser) Parse(data []byte, trailer bool) (chunk, extra []byte, err error) {
	var offset int64

	switch c.state {
	case eChunkLength1Char:
		goto chunkLength1Char
	case eChunkLength:
		goto chunkLength
	case eChunkLengthCR:
		goto chunkLengthCR
	case eChunkLengthCRLF:
		goto chunkLengthCRLF
	case eChunkBody:
		goto chunkBody
	case eChunkBodyEnd:
		goto chunkBodyEnd
	case eChunkBodyCR:
		goto chunkBodyCR
	case eChunkBodyCRLF:
		goto chunkBodyCRLF
	case eLastChunkCR:
		goto lastChunkCR
	case eFooter:
		goto footer
	case eFooterCR:
		goto footerCR
	case eFooterCRLF:
		goto footerCRLF
	case eFooterCRLFCR:
		goto footerCRLFCR
	default:
		panic(fmt.Sprintf("BUG: unknown state: %v", c.state))
	}

chunkLength1Char:
	if !isHex(data[offset]) {
		return nil, nil, status.ErrBadRequest
	}

	c.chunkLength = int64(unHex(data[offset]))
	offset++
	c.state = eChunkLength
	goto chunkLength

chunkLength:
	for ; offset < int64(len(data)); offset++ {
		switch data[offset] {
		case '\r':
			offset++
			c.state = eChunkLengthCR
			goto chunkLengthCR
		case '\n':
			offset++
			c.state = eChunkLengthCRLF
			goto chunkLengthCRLF
		default:
			if !isHex(data[offset]) {
				return nil, nil, status.ErrBadRequest
			}

			c.chunkLength = (c.chunkLength << 4) | int64(unHex(data[offset]))
			if c.chunkLength > c.settings.MaxChunkSize {
				return nil, nil, status.ErrTooLarge
			}
		}
	}

	return nil, nil, nil

chunkLengthCR:
	if offset >= int64(len(data)) {
		return nil, nil, nil
	}

	switch data[offset] {
	case '\n':
		offset++
		c.state = eChunkLengthCRLF
		goto chunkLengthCRLF
	default:
		return nil, nil, status.ErrBadRequest
	}

chunkLengthCRLF:
	if offset >= int64(len(data)) {
		return nil, nil, nil
	}

	switch c.chunkLength {
	case 0:
		switch data[offset] {
		case '\r':
			offset++
			c.state = eLastChunkCR
			goto lastChunkCR
		case '\n':
			c.state = eChunkLength1Char

			return nil, data[offset+1:], io.EOF
		default:
			if !trailer {
				return nil, nil, status.ErrBadRequest
			}

			offset++
			c.state = eFooter
			goto footer
		}
	default:
		c.state = eChunkBody
		goto chunkBody
	}

chunkBody:
	if offset >= int64(len(data)) {
		return nil, nil, nil
	}

	if int64(len(data[offset:])) > c.chunkLength {
		c.state = eChunkBodyEnd

		return data[offset : offset+c.chunkLength], data[offset+c.chunkLength:], nil
	}

	c.chunkLength -= int64(len(data[offset:]))

	return data[offset:], nil, nil

chunkBodyEnd:
	if offset >= int64(len(data)) {
		return nil, nil, nil
	}

	switch data[offset] {
	case '\r':
		offset++
		c.state = eChunkBodyCR
		goto chunkBodyCR
	case '\n':
		offset++
		c.state = eChunkBodyCRLF
		goto chunkBodyCRLF
	default:
		return nil, nil, status.ErrBadRequest
	}

chunkBodyCR:
	if offset >= int64(len(data)) {
		return nil, nil, nil
	}

	switch data[offset] {
	case '\n':
		offset++
		c.state = eChunkBodyCRLF
		goto chunkBodyCRLF
	default:
		return nil, nil, status.ErrBadRequest
	}

chunkBodyCRLF:
	if offset >= int64(len(data)) {
		return nil, nil, nil
	}

	switch data[offset] {
	case '\r':
		offset++
		c.state = eLastChunkCR
		goto lastChunkCR
	case '\n':
		if !trailer {
			c.state = eChunkLength1Char

			return nil, data[offset+1:], io.EOF
		}

		offset++
		c.state = eFooter
		goto footer
	default:
		c.chunkLength = int64(unHex(data[offset]))
		if c.chunkLength > c.settings.MaxChunkSize {
			return nil, nil, status.ErrTooLarge
		}

		offset++
		c.state = eChunkLength
		goto chunkLength
	}

lastChunkCR:
	if offset >= int64(len(data)) {
		return nil, nil, nil
	}

	switch data[offset] {
	case '\n':
		if !trailer {
			c.state = eChunkLength1Char

			return nil, data[offset+1:], io.EOF
		}

		offset++
		c.state = eFooter
		goto footer
	default:
		return nil, nil, status.ErrBadRequest
	}

footer:
	for ; offset < int64(len(data)); offset++ {
		switch data[offset] {
		case '\r':
			offset++
			c.state = eFooterCR
			goto footerCR
		case '\n':
			offset++
			c.state = eFooterCRLF
			goto footerCRLF
		}
	}

	return nil, nil, nil

footerCR:
	if offset >= int64(len(data)) {
		return nil, nil, nil
	}

	switch data[offset] {
	case '\n':
		offset++
		c.state = eFooterCRLF
		goto footerCRLF
	default:
		return nil, nil, status.ErrBadRequest
	}

footerCRLF:
	if offset >= int64(len(data)) {
		return nil, nil, nil
	}

	switch data[offset] {
	case '\r':
		offset++
		c.state = eFooterCRLFCR
		goto footerCRLFCR
	case '\n':
		c.state = eChunkLength1Char

		return nil, data[offset+1:], io.EOF
	default:
		offset++
		c.state = eFooter
		goto footer
	}

footerCRLFCR:
	if offset >= int64(len(data)) {
		return nil, nil, nil
	}

	switch data[offset] {
	case '\n':
		c.state = eChunkLength1Char

		return nil, data[offset+1:], io.EOF
	default:
		return nil, nil, status.ErrBadRequest
	}
}
