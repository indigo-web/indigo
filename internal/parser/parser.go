package parser

// HTTPRequestsParser is a general interface for every http parser
// Currently only http1 parser is presented, but hope in future
// there are will be http2 (at least) parser
type HTTPRequestsParser interface {
	Parse(b []byte) (state RequestState, extra []byte, err error)
	Release()
}
