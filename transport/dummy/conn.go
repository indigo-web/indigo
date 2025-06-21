package dummy

import (
	"io"
	"net"
	"time"
)

type Conn struct {
	Data []byte
	nop  bool
}

func (c *Conn) Read(b []byte) (n int, err error) {
	return 0, io.EOF
}

func (c *Conn) Write(b []byte) (n int, err error) {
	if !c.nop {
		c.Data = append(c.Data, b...)
	}

	return len(b), nil
}

func (c *Conn) Close() error {
	return nil
}

func (c *Conn) LocalAddr() net.Addr {
	return nil
}

func (c *Conn) RemoteAddr() net.Addr {
	return nil
}

func (c *Conn) SetDeadline(t time.Time) error {
	return nil
}

func (c *Conn) SetReadDeadline(t time.Time) error {
	return nil
}

func (c *Conn) SetWriteDeadline(t time.Time) error {
	return nil
}

func (c *Conn) Nop() *Conn {
	c.nop = true
	return c
}
