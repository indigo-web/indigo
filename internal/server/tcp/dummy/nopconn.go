package dummy

import (
	"net"
	"time"
)

// nopConn is implemented in testing purposes, as passing a nil connection to request constructor
// will cause a nil dereference panic
type nopConn struct {
}

func NewNopConn() net.Conn {
	return nopConn{}
}

func (nopConn) Read([]byte) (n int, err error) {
	return
}

func (nopConn) Write([]byte) (n int, err error) {
	return
}

func (nopConn) Close() error {
	return nil
}

func (nopConn) LocalAddr() net.Addr {
	return nil
}

func (nopConn) RemoteAddr() net.Addr {
	return nil
}

func (nopConn) SetDeadline(time.Time) error {
	return nil
}

func (nopConn) SetReadDeadline(time.Time) error {
	return nil
}

func (nopConn) SetWriteDeadline(time.Time) error {
	return nil
}
