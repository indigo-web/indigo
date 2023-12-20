package dummy

import (
	"github.com/indigo-web/indigo/internal/server/tcp"
)

import (
	"io"
	"net"
)

// CircularClient is a client that on every read-operation returns the same data as it
// was initialised with. This is used mainly for benchmarking
type CircularClient struct {
	data            [][]byte
	tmp             []byte
	pointer         int
	closed, oneTime bool
}

func NewCircularClient(data ...[]byte) *CircularClient {
	return &CircularClient{
		data:    data,
		pointer: -1,
	}
}

func (c *CircularClient) Read() (data []byte, err error) {
	if c.closed {
		return nil, io.EOF
	}

	if c.oneTime {
		c.closed = true
	}

	if len(c.tmp) > 0 {
		data, c.tmp = c.tmp, nil

		return data, nil
	}

	c.pointer++

	if c.pointer >= len(c.data) {
		c.pointer = 0
	}

	return c.data[c.pointer], nil
}

func (c *CircularClient) Unread(takeback []byte) {
	c.tmp = takeback
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

func NewNopClient() tcp.Client {
	return NewCircularClient(nil)
}

type SinkholeWriter struct {
	Data []byte
}

func NewSinkholeWriter() *SinkholeWriter {
	return new(SinkholeWriter)
}

func (s *SinkholeWriter) Write(b []byte) error {
	s.Data = append(s.Data, b...)
	return nil
}
