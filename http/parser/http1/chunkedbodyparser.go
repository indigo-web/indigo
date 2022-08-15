package http1

import (
	"indigo/errors"
	"indigo/internal"
	"indigo/settings"
)

type chunkedBodyParser struct {
	state   chunkedBodyParserState
	gateway *internal.BodyGateway

	settings    settings.Settings
	chunkLength uint32
	bodyOffset  int
}

func newChunkedBodyParser(gateway *internal.BodyGateway, settings settings.Settings) chunkedBodyParser {
	return chunkedBodyParser{
		state:    eChunkLength1Char,
		gateway:  gateway,
		settings: settings,
	}
}

func (c *chunkedBodyParser) Parse(data []byte) (done bool, extra []byte, err error) {
	for i := range data {
		switch c.state {
		case eChunkLength1Char:
			if !isHex(data[i]) {
				return true, nil, errors.ErrBadRequest
			}

			c.chunkLength = uint32(unHex(data[i]))
			c.state = eChunkLength
		case eChunkLength:
			switch data[i] {
			case '\r':
				c.state = eChunkLengthCR
			case '\n':
				c.state = eChunkLengthCRLF
			default:
				if !isHex(data[i]) {
					return true, nil, errors.ErrBadRequest
				}

				c.chunkLength = (c.chunkLength << 4) | uint32(unHex(data[i]))
				if c.chunkLength > c.settings.BodyChunkSize.Maximal {
					return true, nil, errors.ErrRequestEntityTooLarge
				}
			}
		case eChunkLengthCR:
			switch data[i] {
			case '\n':
				c.state = eChunkLengthCRLF
			default:
				return true, nil, errors.ErrBadRequest
			}
		case eChunkLengthCRLF:
			if c.chunkLength == 0 {
				switch data[i] {
				case '\r':
					c.state = eLastChunkCR
				case '\n':
					c.state = eChunkLength1Char

					return true, data[i+1:], nil
				default:
					return true, nil, errors.ErrBadRequest
				}
				continue
			}

			c.bodyOffset = i
			c.state = eChunkBody
		case eChunkBody:
			c.chunkLength--

			if c.chunkLength == 0 {
				c.gateway.Data <- data[c.bodyOffset:i]
				<-c.gateway.Data
				if c.gateway.Err != nil {
					return true, nil, c.gateway.Err
				}

				switch data[i] {
				case '\r':
					c.state = eChunkBodyCR
				case '\n':
					c.state = eChunkBodyCRLF
				default:
					return true, nil, errors.ErrBadRequest
				}
			}
		case eChunkBodyCR:
			switch data[i] {
			case '\n':
				c.state = eChunkBodyCRLF
			default:
				return true, nil, errors.ErrBadRequest
			}
		case eChunkBodyCRLF:
			switch data[i] {
			case '\r':
				c.state = eLastChunkCR
			case '\n':
				c.state = eChunkLength1Char

				return true, data[i+1:], nil
			default:
				c.chunkLength = uint32(unHex(data[i]))
				if c.chunkLength > c.settings.BodyChunkSize.Maximal {
					return true, nil, errors.ErrRequestEntityTooLarge
				}

				c.state = eChunkLength
			}
		case eLastChunkCR:
			switch data[i] {
			case '\n':
				c.state = eChunkLength1Char

				return true, data[i+1:], nil
			default:
				return true, nil, errors.ErrBadRequest
			}
		}
	}

	if c.state == eChunkBody {
		c.gateway.Data <- data[c.bodyOffset:]
		<-c.gateway.Data
		if c.gateway.Err != nil {
			return true, nil, c.gateway.Err
		}

		c.bodyOffset = 0
	}

	return false, nil, nil
}
