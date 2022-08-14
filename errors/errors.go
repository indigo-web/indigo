package errors

import (
	"errors"
)

var (
	ErrBadRequest            = errors.New("bad request")
	ErrRequestEntityTooLarge = errors.New("request entity too large")
	ErrURLTooLong            = errors.New("request URI too long")
	ErrURLDecoding           = errors.New("invalid url encoding")
	ErrUnsupportedProtocol   = errors.New("protocol is not supported")

	ErrCloseConnection = errors.New("internal error as a signal")
)
