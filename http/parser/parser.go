package parser

type HTTPRequestsParser interface {
	Parse(b []byte) (state RequestState, extra []byte, err error)
	FinalizeBody()
}
