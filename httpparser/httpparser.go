package httpparser

import (
	"github.com/scott-ainsworth/go-ascii"
	"indigo/errors"
	"indigo/http"
	"indigo/types"
)

type Parser interface {
	Parse(requestStruct *types.Request, data []byte) (done bool, extra []byte, err error)
}

var (
	contentLength    = []byte("content-length")
	transferEncoding = []byte("transfer-encoding")
	connection       = []byte("connection")
	chunked          = []byte("chunked")
	closeConnection  = []byte("close")
)

type HTTPRequestsParser interface {
	Parse([]byte) (done bool, extra []byte, err error)
	Clear()
}

type httpRequestParser struct {
	request  *types.Request
	onBody   types.BodyWriter
	settings Settings

	state            parsingState
	headerValueBegin uint8
	headersBuffer    []byte
	infoLineBuffer   []byte
	infoLineOffset   uint16

	bodyBytesLeft int

	closeConnection bool
	isChunked       bool
	chunksParser    *chunkedBodyParser
}

func NewHTTPParser(request *types.Request, writeBody types.BodyWriter, settings Settings) HTTPRequestsParser {
	settings = PrepareSettings(settings)

	return &httpRequestParser{
		request:        request,
		onBody:         writeBody,
		settings:       settings,
		headersBuffer:  settings.HeadersBuffer,
		infoLineBuffer: settings.InfoLineBuffer,
		chunksParser:   NewChunkedBodyParser(writeBody, settings.MaxChunkSize),
		state:          method,
	}
}

func (p *httpRequestParser) Clear() {
	p.state = method
	p.isChunked = false
	p.headersBuffer = p.headersBuffer[:0]
	p.infoLineBuffer = p.infoLineBuffer[:0]
	p.infoLineOffset = 0
}

/*
	This parser is absolutely stand-alone. It's like a separated sub-system in every
	server, because everything you need is just to feed it
*/
func (p *httpRequestParser) Parse(data []byte) (done bool, extra []byte, err error) {
	if len(data) == 0 {
		if p.closeConnection {
			p.die()
			// to let server know that we received everything, and it's time to close the connection
			return true, nil, errors.ErrConnectionClosed
		}

		return false, nil, nil
	}

	switch p.state {
	case dead:
		return true, nil, errors.ErrParserIsDead
	case messageBegin:
		p.state = method
	case body:
		done, extra, err = p.pushBodyPiece(data)

		if err != nil {
			p.die()

			return true, extra, err
		}

		if done {
			p.Clear()
		}

		return done, extra, nil
	}

	for i := 0; i < len(data); i++ {
		switch p.state {
		case method:
			if data[i] == ' ' {
				method := http.GetMethod(p.infoLineBuffer)
				if method == 0 {
					p.die()

					return true, nil, errors.ErrInvalidMethod
				}

				p.request.Method = method
				p.infoLineOffset = uint16(len(p.infoLineBuffer))
				p.state = path
				break
			}

			p.infoLineBuffer = append(p.infoLineBuffer, data[i])

			if len(p.infoLineBuffer) > MaxMethodLength {
				p.die()

				return true, nil, errors.ErrInvalidMethod
			}
		case path:
			if data[i] == ' ' {
				if uint16(len(p.infoLineBuffer)) == p.infoLineOffset {
					p.die()

					return true, nil, errors.ErrInvalidPath
				}

				p.request.Path = p.infoLineBuffer[p.infoLineOffset:]

				p.infoLineOffset += uint16(len(p.infoLineBuffer[p.infoLineOffset:]))
				p.state = protocol
				continue
			} else if !ascii.IsPrint(data[i]) {
				p.die()

				return true, nil, errors.ErrInvalidPath
			}

			p.infoLineBuffer = append(p.infoLineBuffer, data[i])

			if uint16(len(p.infoLineBuffer[p.infoLineOffset:])) > p.settings.MaxPathLength {
				p.die()

				return true, nil, errors.ErrBufferOverflow
			}
		case protocol:
			switch data[i] {
			case '\r':
				p.state = protocolCR
			case '\n':
				p.state = protocolLF
			default:
				p.infoLineBuffer = append(p.infoLineBuffer, data[i])

				if len(p.infoLineBuffer[p.infoLineOffset:]) > MaxProtocolLength {
					p.die()

					return true, nil, errors.ErrBufferOverflow
				}
			}
		case protocolCR:
			if data[i] != '\n' {
				p.die()

				return true, nil, errors.ErrRequestSyntaxError
			}

			p.state = protocolLF
		case protocolLF:
			proto, ok := http.NewProtocol(p.infoLineBuffer[p.infoLineOffset:])
			if !ok {
				p.die()

				return true, nil, errors.ErrProtocolNotSupported
			}

			p.request.Protocol = *proto

			if data[i] == '\r' {
				p.state = headerValueDoubleCR
				break
			} else if data[i] == '\n' {
				p.Clear()

				return true, data[i+1:], nil
			} else if !ascii.IsPrint(data[i]) || data[i] == ':' {
				p.die()

				return true, nil, errors.ErrInvalidHeader
			}

			p.headersBuffer = append(p.headersBuffer, data[i])
			p.state = headerKey
		case headerKey:
			if data[i] == ':' {
				p.state = headerColon
				p.headerValueBegin = uint8(len(p.headersBuffer))
				break
			} else if !ascii.IsPrint(data[i]) {
				p.die()

				return true, nil, errors.ErrInvalidHeader
			}

			p.headersBuffer = append(p.headersBuffer, data[i])

			if uint8(len(p.headersBuffer)) >= p.settings.MaxHeaderLength {
				p.die()

				return true, nil, errors.ErrBufferOverflow
			}
		case headerColon:
			p.state = headerValue

			if !ascii.IsPrint(data[i]) {
				p.die()

				return true, nil, errors.ErrInvalidHeader
			}

			if data[i] != ' ' {
				p.headersBuffer = append(p.headersBuffer, data[i])
			}
		case headerValue:
			switch data[i] {
			case '\r':
				p.state = headerValueCR
			case '\n':
				p.state = headerValueLF
			default:
				if !ascii.IsPrint(data[i]) {
					p.die()

					return true, nil, errors.ErrInvalidHeader
				}

				p.headersBuffer = append(p.headersBuffer, data[i])

				if uint16(len(p.headersBuffer)) > p.settings.MaxHeaderValueLength {
					p.die()

					return true, nil, errors.ErrBufferOverflow
				}
			}
		case headerValueCR:
			if data[i] != '\n' {
				p.die()

				return true, nil, errors.ErrRequestSyntaxError
			}

			p.state = headerValueLF
		case headerValueLF:
			key, value := p.headersBuffer[:p.headerValueBegin], p.headersBuffer[p.headerValueBegin:]
			p.request.Headers.Set(key, value)

			switch len(key) {
			case len(contentLength):
				good := true

				for j, character := range contentLength {
					if character != (key[j] | 0x20) {
						good = false
						break
					}
				}

				if good {
					if p.bodyBytesLeft, err = parseUint(value); err != nil {
						p.die()

						return true, nil, errors.ErrInvalidContentLength
					}
				}
			case len(transferEncoding):
				good := true

				for j, character := range transferEncoding {
					if character != (key[j] | 0x20) {
						good = false
						break
					}
				}

				if good {
					// TODO: maybe, there are some more transfer encodings I must support?
					p.isChunked = EqualFold(chunked, value)
				}
			case len(connection):
				good := true

				for j, character := range connection {
					if character != (key[j] | 0x20) {
						good = false
						break
					}
				}

				if good {
					p.closeConnection = EqualFold(closeConnection, value)
				}
			}

			switch data[i] {
			case '\r':
				p.state = headerValueDoubleCR
			case '\n':
				if p.closeConnection {
					p.state = bodyConnectionClose
					// anyway in case of empty byte data it will stop parsing, so it's safe
					// but also keeps amount of body bytes limited
					p.bodyBytesLeft = int(p.settings.MaxBodyLength)
					break
				} else if p.bodyBytesLeft == 0 && !p.isChunked {
					p.Clear()

					return true, data[i+1:], nil
				}

				p.state = body
			default:
				p.headersBuffer = append(p.headersBuffer[:0], data[i])
				p.state = headerKey
			}
		case headerValueDoubleCR:
			if data[i] != '\n' {
				p.die()

				return true, nil, errors.ErrRequestSyntaxError
			} else if p.closeConnection {
				p.state = bodyConnectionClose
				p.bodyBytesLeft = int(p.settings.MaxBodyLength)
				break
			} else if p.bodyBytesLeft == 0 && !p.isChunked {
				p.Clear()

				return true, data[i+1:], nil
			}

			p.state = body
		case body:
			done, extra, err = p.pushBodyPiece(data[i:])

			if err != nil {
				p.die()
			} else if done {
				p.Clear()
			}

			return done, extra, err
		case bodyConnectionClose:
			p.bodyBytesLeft -= len(data[i:])

			if p.bodyBytesLeft < 0 {
				p.die()

				return true, nil, errors.ErrBodyTooBig
			}

			p.onBody(data[i:])

			return false, nil, nil
		}
	}

	return false, nil, nil
}

func (p *httpRequestParser) die() {
	p.state = dead
	// anyway we don't need them anymore
	p.headersBuffer = nil
	p.infoLineBuffer = nil
}

func (p *httpRequestParser) pushBodyPiece(data []byte) (done bool, extra []byte, err error) {
	if p.isChunked {
		done, extra, err = p.chunksParser.Feed(data)

		return done, extra, err
	}

	dataLen := len(data)

	if p.bodyBytesLeft > dataLen {
		p.onBody(data)
		p.bodyBytesLeft -= dataLen

		return false, nil, nil
	}

	if p.bodyBytesLeft <= 0 {
		// already?? Looks like a bug
		return true, data, nil
	}

	p.onBody(data[:p.bodyBytesLeft])

	return true, data[p.bodyBytesLeft:], nil
}

func EqualFold(sample, data []byte) bool {
	/*
		Works only for ascii!
	*/

	if len(sample) != len(data) {
		return false
	}

	for i, char := range sample {
		if char != (data[i] | 0x20) {
			return false
		}
	}

	return true
}
