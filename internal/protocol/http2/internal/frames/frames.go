package frames

import "github.com/indigo-web/indigo/internal/protocol/http2/internal/flags"

//go:generate stringer -type=Type -output=frames_string.go
type Type byte

const (
	Data         Type = 0x00
	Headers      Type = 0x01
	Priority     Type = 0x02
	RstStream    Type = 0x03
	Settings     Type = 0x04
	PushPromise  Type = 0x05
	Ping         Type = 0x06
	GoAway       Type = 0x07
	WindowUpdate Type = 0x08
	Continuation Type = 0x09
	Origin       Type = 0x0c
)

type Frame struct {
	Length   uint32
	Type     Type
	Flags    flags.Flag
	StreamID uint32
}
