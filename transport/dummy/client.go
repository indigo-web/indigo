package dummy

import (
	"github.com/indigo-web/indigo/transport"
	"io"
	"net"
)

var _ transport.Client = new(CircularClient)

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
		pointer: 0,
	}
}

func (c *CircularClient) Read() (data []byte, err error) {
	if c.closed {
		return nil, io.EOF
	}

	if len(c.tmp) > 0 {
		data, c.tmp = c.tmp, nil

		return data, nil
	}

	if c.pointer >= len(c.data) {
		if c.oneTime {
			c.closed = true
			return nil, io.EOF
		}

		c.pointer = 0
	}

	piece := c.data[c.pointer]
	c.pointer++

	return piece, nil
}

func (c *CircularClient) Fetch() (data []byte, err error) {
	return c.Read()
}

func (c *CircularClient) Pushback(takeback []byte) {
	c.tmp = takeback
}

func (*CircularClient) Write(p []byte) (int, error) {
	return len(p), nil
}

func (c *CircularClient) Conn() net.Conn {
	return nil
}

func (*CircularClient) Remote() net.Addr {
	return nil
}

func (c *CircularClient) Close() error {
	c.closed = true
	return nil
}

func (c *CircularClient) OneTime() *CircularClient {
	c.oneTime = true
	return c
}

func NewNopClient() transport.Client {
	return NewCircularClient(nil)
}

type SinkholeWriter struct {
	Data []byte
}

func NewSinkholeWriter() *SinkholeWriter {
	return new(SinkholeWriter)
}

func (s *SinkholeWriter) Write(b []byte) (int, error) {
	s.Data = append(s.Data, b...)
	return len(b), nil
}
