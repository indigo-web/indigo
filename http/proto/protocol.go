package proto

import "github.com/flrdv/uf"

type Protocol uint8

const (
	Unknown Protocol = 0
	HTTP10  Protocol = 1 << iota
	HTTP11
	HTTP2

	HTTP1 = HTTP10 | HTTP11
)

func (p Protocol) String() string {
	lut := [...]string{HTTP10: "HTTP/1.0", HTTP11: "HTTP/1.1", HTTP2: "HTTP/2"}
	if int(p) >= len(lut) {
		return ""
	}

	return lut[p]
}

var majorMinorVersionLUT = [10][10]Protocol{
	1: {0: HTTP10, 1: HTTP11},
	2: {0: HTTP2},
}

func FromBytes(raw []byte) Protocol {
	const (
		protoTokenLength   = len("HTTP/x.x")
		majorVersionOffset = len("HTTP/x") - 1
		minorVersionOffset = len("HTTP/x.x") - 1
		httpScheme         = "HTTP/"
	)

	if len(raw) != protoTokenLength || uf.B2S(raw[:majorVersionOffset]) != httpScheme {
		return Unknown
	}

	return Parse(raw[majorVersionOffset]-'0', raw[minorVersionOffset]-'0')
}

func Parse(major, minor uint8) Protocol {
	if major > 9 || minor > 9 {
		return Unknown
	}

	return majorMinorVersionLUT[major][minor]
}
