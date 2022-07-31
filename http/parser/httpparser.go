package parser

import (
	"github.com/scott-ainsworth/go-ascii"
	"indigo/errors"
	"indigo/http"
	"indigo/internal"
	"indigo/settings"
	"indigo/types"
	"io"
)

const (
	maxMethodLength   = len("CONNECT")
	maxProtocolLength = len("HTTP/1.1")
)

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
	request *types.Request
	pipe    internal.Pipe

	settings settings.Settings

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

func NewHTTPParser(request *types.Request, pipe internal.Pipe, settings settings.Settings) HTTPRequestsParser {
	return &httpRequestParser{
		request:        request,
		pipe:           pipe,
		settings:       settings,
		headersBuffer:  make([]byte, 0, settings.DefaultHeadersBuffSize),
		infoLineBuffer: make([]byte, 0, settings.DefaultInfoLineBuffSize),
		chunksParser:   NewChunkedBodyParser(pipe, uint(settings.MaxBodyChunkLength)),
		state:          eMethod,
	}
}

func (p *httpRequestParser) Clear() {
	p.state = eMethod
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
			p.pipe.WriteErr(io.EOF)
			// to let server know that we received everything, and it's time to close the connection
			return true, nil, errors.ErrConnectionClosed
		}

		return false, nil, nil
	}

	switch p.state {
	case eDead:
		return true, nil, errors.ErrParserIsDead
	case eMessageBegin:
		p.state = eMethod
	case eBody:
		done, extra, err = p.pushBodyPiece(data)
		if err != nil {
			p.die()
			// do not write err because the only thing can return error is chunked
			// body parser. But chunked body parser anyway writes error by its own.
			// So we don't care about it

			return true, extra, err
		} else if done {
			if !p.isChunked {
				p.pipe.WriteErr(io.EOF)
			}

			p.Clear()
		}

		return false, extra, nil
	case eBodyConnectionClose:
		p.bodyBytesLeft -= len(data)
		if p.bodyBytesLeft < 0 {
			p.die()
			p.pipe.WriteErr(errors.ErrBodyTooBig)

			return true, nil, errors.ErrBodyTooBig
		}

		p.pipe.Write(data)

		return false, nil, nil
	}

	for i := 0; i < len(data); i++ {
		switch p.state {
		case eMethod:
			if data[i] == ' ' {
				method := http.Bytes2Method(p.infoLineBuffer)
				if method == http.UNKNOWN {
					p.die()

					return true, nil, errors.ErrInvalidMethod
				}

				p.request.Method = method
				p.infoLineOffset = uint16(len(p.infoLineBuffer))
				p.state = ePath
				break
			}

			p.infoLineBuffer = append(p.infoLineBuffer, data[i])

			if len(p.infoLineBuffer) > maxMethodLength {
				p.die()

				return true, nil, errors.ErrInvalidMethod
			}
		case ePath:
			if data[i] == ' ' {
				if uint16(len(p.infoLineBuffer)) == p.infoLineOffset {
					p.die()

					return true, nil, errors.ErrInvalidPath
				}

				p.request.Path = p.infoLineBuffer[p.infoLineOffset:]

				p.infoLineOffset += uint16(len(p.infoLineBuffer[p.infoLineOffset:]))
				p.state = eProtocol
				continue
			} else if !ascii.IsPrint(data[i]) {
				p.die()

				return true, nil, errors.ErrInvalidPath
			}

			p.infoLineBuffer = append(p.infoLineBuffer, data[i])

			if uint16(len(p.infoLineBuffer[p.infoLineOffset:])) > p.settings.MaxURILength {
				p.die()

				return true, nil, errors.ErrBufferOverflow
			}
		case eProtocol:
			switch data[i] {
			case '\r':
				p.state = eProtocolCR
			case '\n':
				p.state = eProtocolLF
			default:
				p.infoLineBuffer = append(p.infoLineBuffer, data[i])

				if len(p.infoLineBuffer[p.infoLineOffset:]) > maxProtocolLength {
					p.die()

					return true, nil, errors.ErrBufferOverflow
				}
			}
		case eProtocolCR:
			if data[i] != '\n' {
				p.die()

				return true, nil, errors.ErrRequestSyntaxError
			}

			p.state = eProtocolLF
		case eProtocolLF:
			proto, ok := http.NewProtocol(p.infoLineBuffer[p.infoLineOffset:])
			if !ok {
				p.die()

				return true, nil, errors.ErrProtocolNotSupported
			}

			p.request.Protocol = *proto

			if data[i] == '\r' {
				p.state = eHeaderValueDoubleCR
				break
			} else if data[i] == '\n' {
				p.Clear()
				p.pipe.WriteErr(io.EOF)

				return true, data[i+1:], nil
			} else if !ascii.IsPrint(data[i]) || data[i] == ':' {
				p.die()

				return true, nil, errors.ErrInvalidHeader
			}

			p.headersBuffer = append(p.headersBuffer, data[i])
			p.state = eHeaderKey
		case eHeaderKey:
			if data[i] == ':' {
				p.state = eHeaderColon
				p.headerValueBegin = uint8(len(p.headersBuffer))
				break
			} else if !ascii.IsPrint(data[i]) {
				p.die()

				return true, nil, errors.ErrInvalidHeader
			}

			p.headersBuffer = append(p.headersBuffer, data[i])

			if uint8(len(p.headersBuffer)) >= p.settings.MaxHeaderKeyLength {
				p.die()

				return true, nil, errors.ErrBufferOverflow
			}
		case eHeaderColon:
			p.state = eHeaderValue

			if !ascii.IsPrint(data[i]) {
				p.die()

				return true, nil, errors.ErrInvalidHeader
			}

			if data[i] != ' ' {
				p.headersBuffer = append(p.headersBuffer, data[i])
			}
		case eHeaderValue:
			switch data[i] {
			case '\r':
				p.state = eHeaderValueCR
			case '\n':
				p.state = eHeaderValueLF
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
		case eHeaderValueCR:
			if data[i] != '\n' {
				p.die()

				return true, nil, errors.ErrRequestSyntaxError
			}

			p.state = eHeaderValueLF
		case eHeaderValueLF:
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
				p.state = eHeaderValueDoubleCR
			case '\n':
				if p.closeConnection {
					p.state = eBodyConnectionClose
					// anyway in case of empty byte data it will stop parsing, so it's safe
					// but also keeps amount of body bytes limited
					p.bodyBytesLeft = int(p.settings.MaxBodyLength)
					return true, data[i+1:], nil
				} else if p.bodyBytesLeft == 0 && !p.isChunked {
					p.Clear()
					p.pipe.WriteErr(io.EOF)

					return true, data[i+1:], nil
				}

				p.state = eBody
				return true, data[i+1:], nil
			default:
				p.headersBuffer = append(p.headersBuffer[:0], data[i])
				p.state = eHeaderKey
			}
		case eHeaderValueDoubleCR:
			if data[i] != '\n' {
				p.die()

				return true, nil, errors.ErrRequestSyntaxError
			} else if p.closeConnection {
				p.state = eBodyConnectionClose
				p.bodyBytesLeft = int(p.settings.MaxBodyLength)

				return true, data[i+1:], nil
			} else if p.bodyBytesLeft == 0 && !p.isChunked {
				p.Clear()
				p.pipe.WriteErr(io.EOF)

				return true, data[i+1:], nil
			}

			p.state = eBody
			return true, data[i+1:], nil
		}
	}

	return false, nil, nil
}

func (p *httpRequestParser) die() {
	p.state = eDead
	// anyway we don't need them anymore
	p.headersBuffer = nil
	p.infoLineBuffer = nil
}

func (p *httpRequestParser) pushBodyPiece(data []byte) (done bool, extra []byte, err error) {
	if p.isChunked {
		return p.chunksParser.Feed(data)
	}

	if p.bodyBytesLeft > len(data) {
		p.pipe.Write(data)
		p.bodyBytesLeft -= len(data)

		return false, nil, nil
	}

	if p.bodyBytesLeft <= 0 {
		p.pipe.WriteErr(io.EOF)
		return true, data, nil
	}

	p.pipe.Write(data[:p.bodyBytesLeft])
	p.pipe.WriteErr(io.EOF)

	return true, data[p.bodyBytesLeft:], nil
}

func EqualFold(sample, data []byte) bool {
	/*
		Works only for ascii characters!
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
