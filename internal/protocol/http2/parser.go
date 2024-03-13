package http2

import (
	"encoding/binary"
	"fmt"
	"github.com/indigo-web/indigo/internal/protocol/http2/internal/flags"
	"github.com/indigo-web/indigo/internal/protocol/http2/internal/frames"
	"github.com/indigo-web/indigo/internal/protocol/http2/internal/settings"
	"github.com/indigo-web/indigo/internal/tcp"
	"github.com/indigo-web/utils/uf"
)

// fun fact: PRI method and SM body are forming together the PRISM, which refers
// to the program used by CIA in order to set up the espionage on the USA biggest
// corporations' users data. In early drafts FOO method and BA body were used instead
const preface = "PRI * HTTP/2.0\r\n\r\nSM\r\n\r\n"

type Parser struct {
	buff   [24]byte
	client tcp.Client
	frame  frames.Frame
	offset uint8
}

func NewParser(client tcp.Client) *Parser {
	return &Parser{
		client: client,
	}
}

func (p *Parser) SkipPreface() error {
	got, err := p.readN(uint8(len(preface)))
	if err != nil {
		return err
	}

	if uf.B2S(got) != preface {
		return fmt.Errorf("bad preface")
	}

	return nil
}

func (p *Parser) Parse() (frame frames.Frame, err error) {
	const frameHeadersOctets = 9
	hdrs, err := p.readN(frameHeadersOctets)
	if err != nil {
		return frame, err
	}

	frame = frames.Frame{
		Length:   uint32(hdrs[2]) | uint32(hdrs[1])<<8 | uint32(hdrs[0])<<16,
		Type:     frames.Type(hdrs[3]),
		Flags:    flags.Flag(hdrs[4]),
		StreamID: binary.BigEndian.Uint32(hdrs[5:9]),
	}

	fmt.Printf(
		"frame (%s): len=%d type=%s flags=%s streamID=%d\n",
		p.client.Remote().String(), frame.Length, frame.Type, frame.Flags, frame.StreamID,
	)

	switch frame.Type {
	case frames.Data:
		panic("no way the data frame to be here")
	case frames.Headers:
	case frames.Priority:
	case frames.RstStream:
	case frames.Settings:
		if frame.Length%6 != 0 {
			return frame, fmt.Errorf("settings payload isn't a multiple of 6")
		}

		for i := uint32(0); i < frame.Length; i += 6 {
			pair, err := p.readN(6)
			if err != nil {
				return frame, err
			}

			key := settings.Setting(binary.BigEndian.Uint16(pair[0:2]))
			value := binary.BigEndian.Uint32(pair[2:6])
			fmt.Printf("setting %s=%d\n", key, value)
		}

		return frame, nil
	case frames.PushPromise:
	case frames.Ping:
	case frames.GoAway:
	case frames.WindowUpdate:
		increment, err := p.readN(4)
		if err != nil {
			return frame, err
		}

		inc := binary.BigEndian.Uint32(increment)
		fmt.Printf("WINDOW_UPDATE: increment=%d\n", inc)
	case frames.Continuation:
	case frames.Origin:
	}

	return frame, nil
}

func (p *Parser) readN(n uint8) ([]byte, error) {
	p.offset = 0
	buff := p.buff[:n]
	for p.offset < n {
		data, err := p.client.Read()
		if err != nil {
			return nil, err
		}

		p.offset += uint8(copy(buff[p.offset:], data))
		p.client.Unread(data[p.offset:])
	}

	return buff, nil
}

func (p *Parser) readByte() (byte, error) {
	b, err := p.readN(1)
	if len(b) > 0 {
		return b[0], err
	}

	return 0, err
}

func (p *Parser) cleanup() {
	p.offset = 0
}
