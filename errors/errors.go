package errors

import (
	"errors"
)

var (
	ErrBadRequest           = errors.New("bad request")
	ErrMethodNotImplemented = errors.New("request method is not supported")
	ErrTooLarge             = errors.New("too large")
	ErrHeaderFieldsTooLarge = errors.New("header fields too large")
	ErrURITooLong           = errors.New("request URI too long")
	ErrURIDecoding          = errors.New("invalid url encoding")
	ErrBadQuery             = errors.New("bad query")
	ErrUnsupportedProtocol  = errors.New("protocol is not supported")
	ErrTooManyHeaders       = errors.New("too much headers")

	ErrCloseConnection = errors.New("internal error as a signal")

	ErrNoSuchKey = errors.New("requested key is not presented")
	ErrRead      = errors.New("body has been already read")

	ErrShutdown   = errors.New("graceful shutdown")
	ErrHijackConn = errors.New("connection hijacking (don't move stay straight)")
)
