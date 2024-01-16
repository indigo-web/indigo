package proto

//go:generate stringer -type=Proto
type Proto uint8

const (
	Unknown Proto = 0
	HTTP09  Proto = 1 << iota
	HTTP10
	HTTP11
	HTTP2

	WebSocket

	HTTP1 = HTTP09 | HTTP10 | HTTP11
)

var (
	http09 = []byte("HTTP/0.9 ")
	http10 = []byte("HTTP/1.0 ")
	http11 = []byte("HTTP/1.1 ")
)

const (
	protoTokenLength   = len("HTTP/x.x")
	majorVersionOffset = len("HTTP/x") - 1
	minorVersionOffset = len("HTTP/x.x") - 1
)

var majorMinorVersionLUT = [10][10]Proto{
	0: {9: HTTP09},
	1: {0: HTTP10, 1: HTTP11},
}

func FromBytes(raw []byte) Proto {
	if len(raw) != protoTokenLength ||
		!(raw[0]|0x20 == 'h' && raw[1]|0x20 == 't' && raw[2]|0x20 == 't' && raw[3]|0x20 == 'p' && raw[4] == '/') {
		return Unknown
	}

	return Parse(raw[majorVersionOffset]-'0', raw[minorVersionOffset]-'0')
}

func Parse(major, minor uint8) Proto {
	if major > 9 || minor > 9 {
		return Unknown
	}

	return majorMinorVersionLUT[major][minor]
}

func ToBytes(proto Proto) []byte {
	switch proto {
	case HTTP09:
		return http09
	case HTTP10:
		return http10
	case HTTP11:
		return http11
	default:
		return nil
	}
}
