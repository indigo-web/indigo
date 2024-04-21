package tcp

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

func (c *client) Unread(b []byte) {
	c.pending = b
}

func (c *client) Conn() net.Conn {
	return c.conn
}

func (c *client) Write(b []byte) error {
	_, err := c.conn.Write(b)

	return err
}

func (c *client) Remote() net.Addr {
	return c.conn.RemoteAddr()
}

func (c *client) Close() error {
	return c.conn.Close()
}
