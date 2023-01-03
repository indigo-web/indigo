package http

type serverState uint8

const (
	eHeadersCompleted serverState = iota + 1
	eProcessed
	eError
	eConnHijack
)
