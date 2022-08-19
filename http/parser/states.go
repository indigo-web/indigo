package parser

// RequestState is just a general state of parser, that notifies
// in case of something new (otherwise Pending is returned)
// for example, it notifies http server whether headers are
// parsed, or request is totally complete. Or RequestCompleted
// as a special case of request with no body presented,
// or even ConnectionClose when Connection: close header was sent,
// and client finally disconnected (empty slice was pushed into the
// parser, http server can know it only from parser)
type RequestState uint8

const (
	Pending RequestState = 1 << iota
	HeadersCompleted
	BodyCompleted
	ConnectionClose
	Error
	RequestCompleted = HeadersCompleted | BodyCompleted
)
