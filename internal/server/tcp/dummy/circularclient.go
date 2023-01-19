package dummy

import (
	"net"

	"github.com/fakefloordiv/indigo/internal/server/tcp"
)

// circularClient is a client that on every read-operation returns the same data as it
//// was initialised with. This is used mainly for benchmarking
type circularClient struct {
	data     [][]byte
	pointer  int
	takeback []byte
}

func NewCircularClient(data ...[]byte) tcp.Client {
	return &circularClient{
		data:    data,
		pointer: -1,
	}
}

func (s *circularClient) Read() ([]byte, error) {
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

func (s *circularClient) Unread(takeback []byte) {
	s.takeback = takeback
}

func (circularClient) Write([]byte) error {
	return nil
}

func (circularClient) Remote() net.Addr {
	return nil
}

func (circularClient) Close() error {
	return nil
}
