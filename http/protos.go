package http

import "indigo/internal"

type ProtocolVersion uint8

const (
	ProtoHTTP09 ProtocolVersion = iota + 1
	ProtoHTTP10
	ProtoHTTP11

	// HTTP2, HTTP3 won't be added until they won't be supported
)

func GetProtocol(proto []byte) ProtocolVersion {
	// yes, dirty. Yes, it's bad. But it's fast!
	internal.ToLowercase(proto)

	switch internal.B2S(proto) {
	case "http/0.9":
		return ProtoHTTP09
	case "http/1.0":
		return ProtoHTTP10
	case "http/1.1":
		return ProtoHTTP11
	}

	return 0
}
