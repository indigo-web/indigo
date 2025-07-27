package dummy

import (
	"io"
	"net"
)

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
