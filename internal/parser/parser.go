package parser

// HTTPRequestsParser is a general interface for every http parser
// Currently only http1 parser is presented, but hope in future
// there are will be http2 (at least) parser
// FinalizeBody() used to notify reader that body is over in case
// no body was presented (see http1/requestsparser.go:FinalizeBody()
// method for more details)
type HTTPRequestsParser interface {
	Parse(b []byte) (state RequestState, extra []byte, err error)
	Release()
	FinalizeBody()
}
