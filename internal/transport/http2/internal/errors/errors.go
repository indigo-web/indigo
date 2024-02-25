package errors

type Error = uint32

const (
	NoError            Error = 0x00
	ProtocolError      Error = 0x01
	InternalError      Error = 0x02
	FlowControlError   Error = 0x03
	SettingsTimeout    Error = 0x04
	StreamClosed       Error = 0x05
	FrameSizeError     Error = 0x06
	RefusedStream      Error = 0x07
	Cancel             Error = 0x08
	CompressionError   Error = 0x09
	ConnectError       Error = 0x0a
	EnhanceYourCalm    Error = 0x0b
	InadequateSecurity Error = 0x0c
	HTTP11Required     Error = 0x0d
)
