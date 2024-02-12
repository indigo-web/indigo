package proto

import "github.com/indigo-web/utils/uf"

//go:generate stringer -type=Proto
type Proto uint8

const (
	Unknown Proto = 0
	HTTP10  Proto = 1 << iota
	HTTP11
	HTTP2

	WebSocket

	HTTP1 = HTTP10 | HTTP11
)

var (
	http10 = []byte("HTTP/1.0 ")
	http11 = []byte("HTTP/1.1 ")
)

const (
	protoTokenLength   = len("HTTP/x.x")
	majorVersionOffset = len("HTTP/x") - 1
	minorVersionOffset = len("HTTP/x.x") - 1
	httpScheme         = "HTTP/"
)

var majorMinorVersionLUT = [10][10]Proto{
	1: {0: HTTP10, 1: HTTP11},
	2: {0: HTTP2},
}

func FromBytes(raw []byte) Proto {
	if len(raw) != protoTokenLength || uf.B2S(raw[:majorVersionOffset]) != httpScheme {
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
	case HTTP10:
		return http10
	case HTTP11:
		return http11
	default:
		return nil
	}
}
