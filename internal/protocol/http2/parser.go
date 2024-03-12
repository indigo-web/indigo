package http2

import (
	"encoding/binary"
	"fmt"
	"github.com/indigo-web/indigo/http/status"
	"github.com/indigo-web/indigo/internal/protocol"
	"github.com/indigo-web/indigo/internal/protocol/http2/internal/flags"
	"github.com/indigo-web/indigo/internal/protocol/http2/internal/frames"
	"github.com/indigo-web/indigo/internal/protocol/http2/internal/settings"
	"github.com/indigo-web/utils/uf"
)

// fun fact: PRI method and SM body are forming together the PRISM, which refers
// to the program used by CIA in order to set up the espionage on the USA biggest
// corporations' users data. In early drafts FOO method and BA body were used instead
const preface = "PRI * HTTP/2.0\r\n\r\nSM\r\n\r\n"

type parserState uint8

const (
	ePreface parserState = iota + 1
	eFrameHeaders
	eSettings
)

type Parser struct {
	state     parserState
	frameType frames.Frame
	flags     flags.Flag
	offset    uint8
	buff      [24]byte
	length    uint32
	streamId  uint32
	settings  settings.Settings
}

func NewParser() *Parser {
	return &Parser{
		state: ePreface,
	}
}

func (p *Parser) Parse(data []byte) (state protocol.RequestState, extra []byte, err error) {
	switch p.state {
	case ePreface:
		goto preface
	case eFrameHeaders:
		goto frameHeaders
	case eSettings:
		goto settings
	default:
		panic(fmt.Sprintf("BUG: http2/parser: unexpected state: %v", p.state))
	}

preface:
	{
		n := copy(p.buff[p.offset:len(preface)], data)
		p.offset += uint8(n)
		if int(p.offset) == len(preface) {
			if uf.B2S(p.buff[:len(preface)]) != preface {
				return protocol.Error, nil, status.ErrBadRequest
			}

			data = data[n:]
			p.offset = 0
			goto frameHeaders
		}

		return protocol.Pending, nil, nil
	}

frameHeaders:
	{
		// frame headers are exactly 9 octets
		const frameHeaders = 9
		n := copy(p.buff[p.offset:frameHeaders], data)
		p.offset += uint8(n)
		if p.offset == frameHeaders {
			headers := p.buff
			p.length = uint32(headers[2]) | uint32(headers[1])<<8 | uint32(headers[0])<<16
			p.frameType = frames.Frame(headers[3])
			p.flags = flags.Flag(headers[4])
			// FIXME: upper-bit must be explicitly set to 0 in order to avoid confusions
			p.streamId = binary.BigEndian.Uint32(headers[5:9])
			data = data[n:]
			p.offset = 0

			fmt.Printf(
				"frame: len=%d type=%s flags=%s streamID=%d\n",
				p.length, p.frameType.String(), p.flags.String(), p.streamId,
			)

			goto dispatchFrame
		}

		p.state = eFrameHeaders
		return protocol.Pending, nil, nil
	}

dispatchFrame:
	switch p.frameType {
	case frames.Settings:
		if p.length%6 != 0 {
			return protocol.Error, nil, status.ErrBadRequest
		}

		goto settings
	}

settings:
	// identifier-value pair in settings are exactly 6 octets each.
	const pair = 6

	for p.length > 0 {
		n := copy(p.buff[p.offset:pair], data)
		data = data[n:]
		p.length -= uint32(n)
		p.offset += uint8(n)
		if p.offset < pair {
			p.state = eSettings
			return protocol.Pending, nil, nil
		}

		p.offset = 0
		key := binary.BigEndian.Uint16(p.buff[:2])
		value := binary.BigEndian.Uint32(p.buff[2:pair])

		if int(key) < len(p.settings) {
			p.settings[key] = value
		}

		fmt.Printf(
			"setting: %s=%d\n",
			settings.Setting(key), value,
		)
	}

	return protocol.HeadersCompleted, nil, nil
}

func (p *Parser) cleanup() {
	p.offset = 0
}
