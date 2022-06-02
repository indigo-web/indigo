package http

type ProtocolVersion uint8

const (
	HTTP09 ProtocolVersion = iota + 1
	HTTP10
	HTTP11

	// HTTP2, HTTP3 won't be added until they won't be supported
)
