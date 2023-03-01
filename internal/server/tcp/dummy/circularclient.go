package dummy

import (
	"github.com/indigo-web/indigo/v2/internal/unreader"
	"io"
	"net"

	"github.com/indigo-web/indigo/v2/internal/server/tcp"
)

// circularClient is a client that on every read-operation returns the same data as it
// was initialised with. This is used mainly for benchmarking
type circularClient struct {
	unreader *unreader.Unreader
	data     [][]byte
	pointer  int
	closed   bool
}

func NewCircularClient(data ...[]byte) tcp.Client {
	return &circularClient{
		unreader: new(unreader.Unreader),
		data:     data,
		pointer:  -1,
	}
}

func (c *circularClient) Read() ([]byte, error) {
	if c.closed {
		return nil, io.EOF
	}

	return c.unreader.PendingOr(func() ([]byte, error) {
		c.pointer++

		if c.pointer == len(c.data) {
			c.pointer = 0
		}

		return c.data[c.pointer], nil
	})
}

func (c *circularClient) Unread(takeback []byte) {
	c.unreader.Unread(takeback)
}

func (*circularClient) Write([]byte) error {
	return nil
}

func (*circularClient) Remote() net.Addr {
	return nil
}

func (c *circularClient) Close() error {
	c.closed = true
	return nil
}
