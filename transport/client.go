package transport

import (
	"net"
	"time"
)

type Client interface {
	Read() ([]byte, error)
	Pushback([]byte)
	Write([]byte) (int, error)
	Conn() net.Conn
	Remote() net.Addr
	Close() error
}

type client struct {
	conn    net.Conn
	buff    []byte
	pending []byte
	timeout time.Duration
}

func NewClient(conn net.Conn, timeout time.Duration, buff []byte) Client {
	return &client{
		buff:    buff,
		conn:    conn,
		timeout: timeout,
	}
}

// Read reads data into the internal buffer and returns a piece of it back. Timeouts are also
// handled automatically.
func (c *client) Read() ([]byte, error) {
	if len(c.pending) > 0 {
		pending := c.pending
		c.pending = nil

		return pending, nil
	}

	if err := c.conn.SetReadDeadline(time.Now().Add(c.timeout)); err != nil {
		return nil, err
	}

	n, err := c.conn.Read(c.buff)
	return c.buff[:n], err
}

// Pending returns data (if any) preserved via Pushback.
func (c *client) Pending() []byte {
	return c.pending
}

// Pushback preserves a chunk of data from previous read for the next read.
func (c *client) Pushback(b []byte) {
	c.pending = b
}

// Conn unwraps the underlying net.Conn.
func (c *client) Conn() net.Conn {
	return c.conn
}

// Write writes data into the underlying connection.
func (c *client) Write(b []byte) (int, error) {
	return c.conn.Write(b)
}

// Remote returns the remote address of the connection.
func (c *client) Remote() net.Addr {
	return c.conn.RemoteAddr()
}

// Close closes the connection.
func (c *client) Close() error {
	return c.conn.Close()
}
