package dummy

import (
	"net"
	"time"
)

// NopConn is used to bypass a connection object, that does absolutely nothing. It exists just
// in order to be passed where written data isn't the point
type NopConn struct{}

func NewNopConn() *NopConn {
	return new(NopConn)
}

func (NopConn) Read([]byte) (n int, err error) {
	return
}

func (NopConn) Write([]byte) (n int, err error) {
	return
}

func (NopConn) Close() error {
	return nil
}

func (NopConn) LocalAddr() net.Addr {
	return nil
}

func (NopConn) RemoteAddr() net.Addr {
	return nil
}

func (NopConn) SetDeadline(time.Time) error {
	return nil
}

func (NopConn) SetReadDeadline(time.Time) error {
	return nil
}

func (NopConn) SetWriteDeadline(time.Time) error {
	return nil
}
