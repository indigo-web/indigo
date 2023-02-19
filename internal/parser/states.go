package parser

// RequestState is a general state of the parser that tells a caller about
// current state of the request. It may be incomplete (Pending), complete
// (HeadersCompleted), and completed with an error (Error). Due to internal
// approaches of body handling, parser does not manage requests bodies.
type RequestState uint8

const (
	Pending RequestState = iota + 1
	HeadersCompleted
	Error
)
