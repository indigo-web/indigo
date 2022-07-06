package http

import "indigo/internal"

type protocolVersion uint8

const (
	protoHTTP09 protocolVersion = iota + 1
	protoHTTP10
	protoHTTP11

	// HTTP2, HTTP3 won't be added until they won't be supported
)

type Protocol struct {
	enum protocolVersion
	raw  []byte
}

func NewProtocol(proto []byte) (*Protocol, bool) {
	// yes, dirty. Yes, it's bad. But it's fast!
	internal.ToLowercase(proto)

	var protoEnum protocolVersion

	switch internal.B2S(proto) {
	case "http/0.9":
		protoEnum = protoHTTP09
	case "http/1.0":
		protoEnum = protoHTTP10
	case "http/1.1":
		protoEnum = protoHTTP11
	default:
		return nil, false
	}

	return &Protocol{
		enum: protoEnum,
		raw:  append(proto, '\r', '\n'),
	}, true
}

func (p Protocol) IsHTTP09() bool {
	return p.enum == protoHTTP09
}

func (p Protocol) IsHTTP10() bool {
	return p.enum == protoHTTP10
}

func (p Protocol) IsHTTP11() bool {
	return p.enum == protoHTTP11
}

func (p Protocol) Raw() []byte {
	return p.raw
}
