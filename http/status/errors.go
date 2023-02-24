package status

import (
	"errors"
)

var (
	ErrBadRequest           = errors.New("bad request")
	ErrMethodNotImplemented = errors.New("request method is not supported")
	ErrTooLarge             = errors.New("too large")
	ErrHeaderFieldsTooLarge = errors.New("header fields too large")
	ErrURITooLong           = errors.New("request URI too long")
	ErrURIDecoding          = errors.New("invalid URI encoding")
	ErrBadQuery             = errors.New("bad URL query")
	ErrUnsupportedProtocol  = errors.New("protocol is not supported")
	ErrUnsupportedEncoding  = errors.New("content encoding is not supported")
	ErrTooManyHeaders       = errors.New("too many headers")

	ErrConnectionTimeout = errors.New("connection timed out")
	ErrCloseConnection   = errors.New("internal error as a signal")

	ErrNotFound            = errors.New("not found")
	ErrMethodNotAllowed    = errors.New("method is not allowed")
	ErrInternalServerError = errors.New("internal server error")

	ErrShutdown = errors.New("graceful shutdown")
)
