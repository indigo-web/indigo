package errors

import (
	"errors"
)

var (
	ErrBadRequest          = errors.New("bad request")
	ErrTooLarge            = errors.New("too large")
	ErrURLTooLong          = errors.New("request URI too long")
	ErrURLDecoding         = errors.New("invalid url encoding")
	ErrBadQuery            = errors.New("bad query")
	ErrUnsupportedProtocol = errors.New("protocol is not supported")
	ErrTooManyHeaders      = errors.New("too much headers")

	ErrCloseConnection = errors.New("internal error as a signal")

	ErrNoSuchKey = errors.New("requested key is not presented")
	ErrRead      = errors.New("body has been already read")
)
