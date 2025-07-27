package dummy

import (
	"github.com/indigo-web/indigo/transport"
	"io"
	"net"
)

var _ transport.Client = new(Client)

// Client returns the same data as it was initialised with on every read, unless set to
// shoot once. It also tracks all the written data, making it thereby a universal mock
// suitable for most of the tests.
type Client struct {
	closed     bool
	once       bool
	journaling bool
	pointer    int
	tmp        []byte
	written    []byte
	data       [][]byte
}

func NewMockClient(data ...[]byte) *Client {
	return &Client{
		data:       data,
		pointer:    0,
		journaling: true,
	}
}

func (c *Client) Read() (data []byte, err error) {
	if c.closed {
		return nil, io.EOF
	}

	if len(c.tmp) > 0 {
		data, c.tmp = c.tmp, nil

		return data, nil
	}

	if c.pointer >= len(c.data) {
		if c.once {
			c.closed = true
			return nil, io.EOF
		}

		c.pointer = 0
	}

	piece := c.data[c.pointer]
	c.pointer++

	return piece, nil
}

func (c *Client) Fetch() (data []byte, err error) {
	return c.Read()
}

func (c *Client) Pushback(takeback []byte) {
	c.tmp = takeback
}

func (c *Client) Write(p []byte) (int, error) {
	if c.journaling {
		c.written = append(c.written, p...)
	}

	return len(p), nil
}

func (c *Client) Conn() net.Conn {
	return new(Conn).Nop()
}

func (*Client) Remote() net.Addr {
	return nil
}

func (c *Client) Close() error {
	c.closed = true
	return nil
}

func (c *Client) Once() *Client {
	c.once = true
	return c
}

func (c *Client) Journaling(flag bool) *Client {
	c.journaling = flag
	return c
}

func (c *Client) Written() string {
	if !c.journaling {
		panic("mock client: cannot access written data: journaling is disabled!")
	}

	return string(c.written)
}
