package parser

// HTTPRequestsParser is a general interface for every http parser
// Currently only http1 parser is presented, but hope in future
// there are will be at least http2 parser
type HTTPRequestsParser interface {
	Parse(b []byte) (state RequestState, extra []byte, err error)
	Release()
}
