package http1

import (
	"github.com/fakefloordiv/indigo/http"
	"github.com/fakefloordiv/indigo/http/encodings"
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
	chunkLength uint32
	bodyOffset  int
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
		decoder = func(b []byte) ([]byte, error) {
			return b, nil
		}
	}

	for i := range data {
		switch c.state {
		case eChunkLength1Char:
			if !isHex(data[i]) {
				return true, nil, http.ErrBadRequest
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
					return true, nil, http.ErrBadRequest
				}

				c.chunkLength = (c.chunkLength << 4) | uint32(unHex(data[i]))
				if c.chunkLength > c.settings.Body.ChunkSize.Maximal {
					return true, nil, http.ErrTooLarge
				}
			}
		case eChunkLengthCR:
			switch data[i] {
			case '\n':
				c.state = eChunkLengthCRLF
			default:
				return true, nil, http.ErrBadRequest
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
					if !trailer {
						return true, nil, http.ErrBadRequest
					}

					c.state = eFooter
				}

				continue
			}

			c.bodyOffset = i
			c.state = eChunkBody
		case eChunkBody:
			c.chunkLength--

			if c.chunkLength == 0 {
				decoded, err := decoder(data[c.bodyOffset:i])
				if err != nil {
					return true, nil, err
				}

				c.gateway.Data <- decoded
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
					return true, nil, http.ErrBadRequest
				}
			}
		case eChunkBodyCR:
			switch data[i] {
			case '\n':
				c.state = eChunkBodyCRLF
			default:
				return true, nil, http.ErrBadRequest
			}
		case eChunkBodyCRLF:
			switch data[i] {
			case '\r':
				c.state = eLastChunkCR
			case '\n':
				if !trailer {
					c.state = eChunkLength1Char

					return true, data[i+1:], nil
				}

				c.state = eFooter
			default:
				c.chunkLength = uint32(unHex(data[i]))
				if c.chunkLength > c.settings.Body.ChunkSize.Maximal {
					return true, nil, http.ErrTooLarge
				}

				c.state = eChunkLength
			}
		case eLastChunkCR:
			switch data[i] {
			case '\n':
				if !trailer {
					c.state = eChunkLength1Char

					return true, data[i+1:], nil
				}

				c.state = eFooter
			default:
				return true, nil, http.ErrBadRequest
			}
		case eFooter:
			switch data[i] {
			case '\r':
				c.state = eFooterCR
			case '\n':
				c.state = eFooterCRLF
			}
		case eFooterCR:
			switch data[i] {
			case '\n':
				c.state = eFooterCRLF
			default:
				return true, nil, http.ErrBadRequest
			}
		case eFooterCRLF:
			switch data[i] {
			case '\r':
				c.state = eFooterCRLFCR
			case '\n':
				c.state = eChunkLength1Char

				return true, data[i+1:], nil
			default:
				c.state = eFooter
			}
		case eFooterCRLFCR:
			switch data[i] {
			case '\n':
				c.state = eChunkLength1Char

				return true, data[i+1:], nil
			default:
				return done, nil, http.ErrBadRequest
			}
		}
	}

	if c.state == eChunkBody {
		decoded, err := decoder(data[c.bodyOffset:])
		if err != nil {
			return true, nil, err
		}

		c.gateway.Data <- decoded
		<-c.gateway.Data
		if c.gateway.Err != nil {
			return true, nil, c.gateway.Err
		}

		c.bodyOffset = 0
	}

	return false, nil, nil
}
