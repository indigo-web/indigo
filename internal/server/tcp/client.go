package tcp

import (
	"github.com/indigo-web/utils/unreader"
	"net"
	"time"
)

type Client interface {
	Read() ([]byte, error)
	Unread([]byte)
	Write([]byte) error
	Remote() net.Addr
	Close() error
}

type client struct {
	conn     net.Conn
	unreader *unreader.Unreader
	buff     []byte
	timeout  time.Duration
}

func NewClient(conn net.Conn, timeout time.Duration, buff []byte) Client {
	return &client{
		unreader: new(unreader.Unreader),
		buff:     buff,
		conn:     conn,
		timeout:  timeout,
	}
}

func (c *client) Read() ([]byte, error) {
	return c.unreader.PendingOr(func() ([]byte, error) {
		if err := c.conn.SetReadDeadline(time.Now().Add(c.timeout)); err != nil {
			return nil, err
		}

		n, err := c.conn.Read(c.buff)

		return c.buff[:n], err
	})
}

func (c *client) Unread(b []byte) {
	c.unreader.Unread(b)
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
