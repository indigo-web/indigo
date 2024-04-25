package proto

import "github.com/indigo-web/utils/uf"

type Proto uint8

const (
	Unknown Proto = 0
	HTTP10  Proto = 1 << iota
	HTTP11
	HTTP2

	HTTP1 = HTTP10 | HTTP11
)

// String returns protocol as a string WITH A TRAILING SPACE
func (p Proto) String() string {
	lut := [...]string{HTTP10: "HTTP/1.0 ", HTTP11: "HTTP/1.1 ", HTTP2: "HTTP/2 "}
	if int(p) >= len(lut) {
		return ""
	}

	return lut[p]
}

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
