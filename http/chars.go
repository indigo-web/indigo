package http

var (
	CR   byte = '\r'
	LF   byte = '\n'
	CRLF      = []byte{CR, LF}
)

var (
	SP      = " "
	COLON   = ":"
	COLONSP = COLON + SP
	COMMA   = ","
)
