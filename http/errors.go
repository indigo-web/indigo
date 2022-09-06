package http

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
	ErrUnsupportedEncoding  = errors.New("content encoding is not supported")
	ErrTooManyHeaders       = errors.New("too much headers")

	ErrCloseConnection = errors.New("internal error as a signal")

	ErrNotFound         = errors.New("not found")
	ErrMethodNotAllowed = errors.New("method is not allowed")

	ErrShutdown   = errors.New("graceful shutdown")
	ErrHijackConn = errors.New("connection hijacking (don't move stay straight)")
)
