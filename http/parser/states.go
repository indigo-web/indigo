package parser

type RequestState uint8

const (
	Pending RequestState = 1 << iota
	HeadersCompleted
	BodyCompleted
	ConnectionClose
	Error
	RequestCompleted = HeadersCompleted | BodyCompleted
)
