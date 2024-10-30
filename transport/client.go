package transport

import (
	"net"
	"time"
)

type Client interface {
	Read() ([]byte, error)
	Unread([]byte)
	Write([]byte) error
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

// Read returns a byte-slice containing data read from the connection. Each calls
// return the same slice with different content, so each call basically overrides
// result of the previous one
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

// Pending returns stored pending data
func (c *client) Pending() []byte {
	return c.pending
}

// Unread stores passed data into the pending buffer. Previous content will
// be lost
func (c *client) Unread(b []byte) {
	c.pending = b
}

// Conn returns the actual connection object
func (c *client) Conn() net.Conn {
	return c.conn
}

// Write writes data into the underlying connection
func (c *client) Write(b []byte) error {
	_, err := c.conn.Write(b)
	return err
}

// Remote returns the remote address of the connection
func (c *client) Remote() net.Addr {
	return c.conn.RemoteAddr()
}

// Close closes the connection
func (c *client) Close() error {
	return c.conn.Close()
}
