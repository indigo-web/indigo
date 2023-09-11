package http1

type parserState int

const (
	eProto parserState = iota + 1
	eCode
	eStatus
	eHeaderKey
	eHeaderKeyCR
	eHeaderSemicolon
	eHeaderValue
)
