package errors

import "errors"

type Error error

var (
	ErrDuplicatedHeader = errors.New("found a duplicated header")

	ErrInvalidMethod        = errors.New("ErrInvalidMethod: invalid method")
	ErrInvalidPath          = errors.New("ErrInvalidPath: path is empty or contains disallowed characters")
	ErrProtocolNotSupported = errors.New("ErrProtocolNotSupported: protocol is not supported")
	ErrInvalidHeader        = errors.New("ErrInvalidHeader: invalid header line")
	ErrBufferOverflow       = errors.New("ErrBufferOverflow: buffer overflow")
	ErrInvalidContentLength = errors.New("ErrInvalidContentLength: invalid value for content-length header")
	ErrRequestSyntaxError   = errors.New("ErrRequestSyntaxError: request syntax error")
	ErrBodyTooBig           = errors.New("ErrBodyTooBig: received too much body before connection closed")

	ErrTooBigChunkSize      = errors.New("ErrTooBigChunkSize: chunk size is too big")
	ErrInvalidChunkSize     = errors.New("ErrInvalidChunkSize: chunk size is invalid hexdecimal value")
	ErrInvalidChunkSplitter = errors.New("ErrInvalidChunkSplitter: invalid splitter")

	ErrConnectionClosed = errors.New("ErrConnectionClosed: connection is closed, body has been received")
	ErrParserIsDead     = errors.New("ErrParserIsDead: BUG: once error occurred, parser cannot be used anymore")
)
