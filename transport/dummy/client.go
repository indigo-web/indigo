package dummy

import (
	"io"
	"net"

	"github.com/indigo-web/indigo/transport"
)

var _ transport.Client = new(Client)

// Client is a full-blown client implementation intended for mock tests, thereby providing relatively
// reach functionality.
type Client struct {
	closed    bool
	loopReads bool
	pointer   int
	conn      *Conn
	pending   []byte
	data      [][]byte
}

func NewMockClient(data ...[]byte) *Client {
	return &Client{
		data: data,
		conn: new(Conn).Nop(),
	}
}

func (c *Client) Read() (data []byte, err error) {
	if len(c.pending) > 0 {
		data, c.pending = c.pending, nil

		return data, nil
	}

	if c.pointer >= len(c.data) {
		if !c.loopReads {
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
	c.pending = takeback
}

func (c *Client) Write(p []byte) (int, error) {
	return c.conn.Write(p)
}

func (c *Client) Written() []byte {
	if c.conn.nop {
		panic("mock client: cannot access written data: journaling is disabled!")
	}

	return c.conn.Data
}

func (c *Client) Conn() net.Conn {
	return c.conn
}

func (*Client) Remote() net.Addr {
	return nil
}

func (c *Client) Close() error {
	c.closed = true
	return nil
}

func (c *Client) Closed() bool {
	return c.closed
}

func (c *Client) Journaling() *Client {
	c.conn.nop = false
	return c
}

// LoopReads disables returning io.EOF on data exhaustion, instead starting from the
// beginning.
func (c *Client) LoopReads() *Client {
	c.loopReads = true
	return c
}

func (c *Client) Reset() {
	c.conn.Data = c.conn.Data[:0]
}

var _ transport.Client = NopClient{}

type NopClient struct{}

func NewNopClient() NopClient {
	return NopClient{}
}

func (n NopClient) Read() ([]byte, error) {
	return nil, io.EOF
}

func (n NopClient) Pushback([]byte) {}

func (n NopClient) Write(b []byte) (int, error) {
	return len(b), nil
}

func (n NopClient) Conn() net.Conn {
	return new(Conn).Nop()
}

func (n NopClient) Remote() net.Addr {
	return nil
}

func (n NopClient) Close() error {
	return nil
}
