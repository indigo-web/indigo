package tcp

import (
	"net"
)

// staticClient is implemented in benchmarking purposes
type staticClient struct {
	data     [][]byte
	pointer  int
	takeback []byte
}

// NewStaticClient is a client that on every read-operation returns the same data as it
// was initialised with. This is used mainly for benchmarking
func NewStaticClient(data ...[]byte) Client {
	return &staticClient{
		data:    data,
		pointer: -1,
	}
}

func (s *staticClient) Read() ([]byte, error) {
	if len(s.takeback) > 0 {
		takeback := s.takeback
		s.takeback = nil

		return takeback, nil
	}

	s.pointer++

	if s.pointer == len(s.data) {
		s.pointer = 0
	}

	return s.data[s.pointer], nil
}

func (s *staticClient) Unread(takeback []byte) {
	s.takeback = takeback
}

func (s staticClient) Write([]byte) error {
	return nil
}

func (s staticClient) Remote() net.Addr {
	return nil
}

func (s staticClient) Close() error {
	return nil
}
