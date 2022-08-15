package server

type serverState uint8

const (
	headersCompleted serverState = iota + 1
	processed
	closeConnection
	badRequest
)
