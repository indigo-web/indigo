package dummy

import (
	"github.com/indigo-web/utils/unreader"
	"io"
	"net"
)

// CircularClient is a client that on every read-operation returns the same data as it
// was initialised with. This is used mainly for benchmarking
type CircularClient struct {
	unreader        *unreader.Unreader
	data            [][]byte
	pointer         int
	closed, oneTime bool
}

func NewCircularClient(data ...[]byte) *CircularClient {
	return &CircularClient{
		unreader: new(unreader.Unreader),
		data:     data,
		pointer:  -1,
	}
}

func (c *CircularClient) Read() ([]byte, error) {
	if c.closed {
		return nil, io.EOF
	}

	if c.oneTime {
		c.closed = true
	}

	return c.unreader.PendingOr(func() ([]byte, error) {
		c.pointer++

		if c.pointer == len(c.data) {
			c.pointer = 0
		}

		return c.data[c.pointer], nil
	})
}

func (c *CircularClient) Unread(takeback []byte) {
	c.unreader.Unread(takeback)
}

func (*CircularClient) Write([]byte) error {
	return nil
}

func (*CircularClient) Remote() net.Addr {
	return nil
}

func (c *CircularClient) Close() error {
	c.closed = true
	return nil
}

func (c *CircularClient) OneTime() {
	c.oneTime = true
}
