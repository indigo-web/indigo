package proto

import (
	"github.com/indigo-web/utils/strcomp"
	"github.com/indigo-web/utils/uf"
)

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
	minimalProtoStringVersion = len("HTTP/x.x")
	httpProtoPrefix           = "HTTP/"
	majorVersionOffset        = len("HTTP/x") - 1
	minorVersionOffset        = len("HTTP/x.x") - 1
)

func FromBytes(raw []byte) Proto {
	if len(raw) != minimalProtoStringVersion ||
		!strcomp.EqualFold(uf.B2S(raw[:len(httpProtoPrefix)]), httpProtoPrefix) {
		return Unknown
	}

	return Parse(raw[majorVersionOffset]-'0', raw[minorVersionOffset]-'0')
}

func Parse(major, minor uint8) Proto {
	switch {
	case major == 1 && minor == 1:
		return HTTP11
	case major == 1 && minor == 0:
		return HTTP10
	case major == 0 && minor == 9:
		return HTTP09
	}

	return Unknown
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
