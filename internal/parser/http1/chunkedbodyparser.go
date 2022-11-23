package http1

import (
	"fmt"
	"github.com/fakefloordiv/indigo/http/encodings"
	"github.com/fakefloordiv/indigo/http/status"
	"github.com/fakefloordiv/indigo/internal/body"
	"github.com/fakefloordiv/indigo/settings"
)

// chunkedBodyParser is a parser for chunked encoded request bodies
// used to encapsulate process of parsing because it's more convenient
// to leave the process here and let main parser parse only http requests
type chunkedBodyParser struct {
	state   chunkedBodyParserState
	gateway *body.Gateway

	settings    settings.Settings
	chunkLength int
}

func newChunkedBodyParser(gateway *body.Gateway, settings settings.Settings) chunkedBodyParser {
	return chunkedBodyParser{
		state:    eChunkLength1Char,
		gateway:  gateway,
		settings: settings,
	}
}

// Parse takes only body as it is. Returns a flag whether parsing is done,
// extra that are extra-bytes related to a next one request, and err if
// occurred
func (c *chunkedBodyParser) Parse(
	data []byte, decoder encodings.DecoderFunc, trailer bool,
) (done bool, extra []byte, err error) {
	if decoder == nil {
		decoder = nopDecoder
	}

	var (
		offset  int
		decoded []byte
	)

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
	}

chunkLength1Char:
	if !isHex(data[offset]) {
		fmt.Println("?")
		return true, nil, status.ErrBadRequest
	}

	c.chunkLength = int(unHex(data[offset]))
	offset++
	c.state = eChunkLength
	goto chunkLength

chunkLength:
	for ; offset < len(data); offset++ {
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
				return true, nil, status.ErrBadRequest
			}

			c.chunkLength = (c.chunkLength << 4) | int(unHex(data[offset]))
			if c.chunkLength > int(c.settings.Body.ChunkSize.Maximal) {
				return true, nil, status.ErrTooLarge
			}
		}
	}

	return false, nil, nil

chunkLengthCR:
	if offset >= len(data) {
		return false, nil, nil
	}

	switch data[offset] {
	case '\n':
		offset++
		c.state = eChunkLengthCRLF
		goto chunkLengthCRLF
	default:
		fmt.Println("???")
		return true, nil, status.ErrBadRequest
	}

chunkLengthCRLF:
	if offset >= len(data) {
		return false, nil, nil
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

			return true, data[offset+1:], nil
		default:
			if !trailer {
				return true, nil, status.ErrBadRequest
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
	if offset >= len(data) {
		return false, nil, nil
	}

	if c.chunkLength <= len(data[offset:]) {
		decoded, err = decoder(data[offset : offset+c.chunkLength])
		offset += c.chunkLength
		c.chunkLength = 0
	} else {
		decoded, err = decoder(data[offset:])
		c.chunkLength -= len(data[offset:])
		offset += len(data[offset:])
	}

	if err != nil {
		return true, nil, err
	}

	c.gateway.Data <- decoded
	<-c.gateway.Data
	if c.gateway.Err != nil {
		return true, nil, c.gateway.Err
	}

	if c.chunkLength == 0 && offset < len(data) {
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
			return true, nil, status.ErrBadRequest
		}
	}

	return false, nil, nil

chunkBodyCR:
	if offset >= len(data) {
		return false, nil, nil
	}

	switch data[offset] {
	case '\n':
		offset++
		c.state = eChunkBodyCRLF
		goto chunkBodyCRLF
	default:
		return true, nil, status.ErrBadRequest
	}

chunkBodyCRLF:
	if offset >= len(data) {
		return false, nil, nil
	}

	switch data[offset] {
	case '\r':
		offset++
		c.state = eLastChunkCR
		goto lastChunkCR
	case '\n':
		if !trailer {
			c.state = eChunkLength1Char

			return true, data[offset+1:], nil
		}

		offset++
		c.state = eFooter
		goto footer
	default:
		c.chunkLength = int(unHex(data[offset]))
		if c.chunkLength > int(c.settings.Body.ChunkSize.Maximal) {
			return true, nil, status.ErrTooLarge
		}

		offset++
		c.state = eChunkLength
		goto chunkLength
	}

lastChunkCR:
	if offset >= len(data) {
		return false, nil, nil
	}

	switch data[offset] {
	case '\n':
		if !trailer {
			c.state = eChunkLength1Char

			return true, data[offset+1:], nil
		}

		offset++
		c.state = eFooter
		goto footer
	default:
		return true, nil, status.ErrBadRequest
	}

footer:
	for ; offset < len(data); offset++ {
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

	return false, nil, nil

footerCR:
	if offset >= len(data) {
		return false, nil, nil
	}

	switch data[offset] {
	case '\n':
		offset++
		c.state = eFooterCRLF
		goto footerCRLF
	default:
		return true, nil, status.ErrBadRequest
	}

footerCRLF:
	if offset >= len(data) {
		return false, nil, nil
	}

	switch data[offset] {
	case '\r':
		offset++
		c.state = eFooterCRLFCR
		goto footerCRLFCR
	case '\n':
		c.state = eChunkLength1Char

		return true, data[offset+1:], nil
	default:
		offset++
		c.state = eFooter
		goto footer
	}

footerCRLFCR:
	if offset >= len(data) {
		return false, nil, nil
	}

	switch data[offset] {
	case '\n':
		c.state = eChunkLength1Char

		return true, data[offset+1:], nil
	default:
		return done, nil, status.ErrBadRequest
	}
}

func nopDecoder(b []byte) ([]byte, error) {
	return b, nil
}
