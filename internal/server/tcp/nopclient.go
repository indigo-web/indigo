package tcp

import (
	"io"
	"net"
)

// nopClient is implemented in testing purposes
type nopClient struct {
	takeback []byte
}

// NewNopClient returns a new nop-client that does nothing except takebacks (this works as intended
// to be)
func NewNopClient() Client {
	return &nopClient{}
}

func (n *nopClient) Read() ([]byte, error) {
	if len(n.takeback) > 0 {
		takeback := n.takeback
		n.takeback = nil

		return takeback, nil
	}

	return nil, io.EOF
}

func (n *nopClient) Unread(takeback []byte) {
	n.takeback = takeback
}

func (nopClient) Write([]byte) error {
	return nil
}

func (nopClient) Remote() net.Addr {
	return &net.TCPAddr{}
}

func (nopClient) Close() error {
	return nil
}
