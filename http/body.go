package http

type OnBodyCallback func([]byte) error

type Body interface {
	Init(*Request)
	Retrieve() ([]byte, error)
	Read([]byte) (n int, err error)
	String() (string, error)
	Bytes() ([]byte, error)
	Callback(cb OnBodyCallback) error
	Reset() error
}
