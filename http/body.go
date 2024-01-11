package http

type OnBodyCallback func([]byte) error

type Body interface {
	// Init is used to initialize the Body instance. MUST NOT be used, as otherwise
	// deadlock may be encountered
	Init(*Request)
	// Retrieve reads request's body from a socket. If there's no data yet, the call will
	// be blocked. It's safe to call the method after whole body was read, as only io.EOF
	// will be returned
	Retrieve() ([]byte, error)
	// Read implements io.Reader interface
	Read([]byte) (n int, err error)
	// String returns the whole request's body at once as a string
	String() (string, error)
	// Bytes returns the whole request's body at once as a byte slice
	Bytes() ([]byte, error)
	// Callback takes a function that'll be called each time as at least 1 byte of request's
	// body is received. The call will be blocked until the whole body won't be processed.
	// When the body is completely processed, the method will silently exit without notifying
	// the passed function anyhow
	Callback(cb OnBodyCallback) error
	// Discard discards the rest of the body
	Discard() error
}
