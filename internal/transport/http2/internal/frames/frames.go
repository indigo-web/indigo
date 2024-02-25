package frames

//go:generate stringer -type=Frame -output=frames_string.go
type Frame byte

const (
	Data         Frame = 0x00
	Headers      Frame = 0x01
	Priority     Frame = 0x02
	RstStream    Frame = 0x03
	Settings     Frame = 0x04
	PushPromise  Frame = 0x05
	Ping         Frame = 0x06
	GoAway       Frame = 0x07
	WindowUpdate Frame = 0x08
	Continuation Frame = 0x09
	Origin       Frame = 0x0c
)
