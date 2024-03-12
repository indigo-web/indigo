package settings

//go:generate stringer -type=Setting -output=settings_string.go
type Setting uint16

const (
	HeaderTableSize      Setting = 0x01
	EnablePush           Setting = 0x02
	MaxConcurrentStreams Setting = 0x03
	InitialWindowSize    Setting = 0x04
	MaxFrameSize         Setting = 0x05
	MaxHeaderListSize    Setting = 0x06
)

type Settings [7]uint32

var Default = Settings{
	HeaderTableSize:      4096,
	EnablePush:           0,
	MaxConcurrentStreams: 100,
	InitialWindowSize:    65535,
	MaxFrameSize:         16384,
	MaxHeaderListSize:    0xffffffff, // basically unlimited
}
