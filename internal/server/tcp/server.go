package tcp

import (
	"github.com/indigo-web/indigo/http/status"
	"net"
	"sync"
)

type OnConn func(net.Conn)

type Server struct {
	sock     net.Listener
	onConn   OnConn
	mu       sync.Mutex
	conns    map[net.Conn]struct{}
	shutdown bool
}

func NewServer(sock net.Listener, onConn OnConn) *Server {
	return &Server{
		sock:   sock,
		onConn: onConn,
		conns:  map[net.Conn]struct{}{},
	}
}

func (s *Server) Start() error {
	wg := new(sync.WaitGroup)

	for {
		conn, err := s.sock.Accept()
		if err != nil {
			wg.Wait()

			if s.shutdown {
				return status.ErrShutdown
			}

			return err
		}

		s.mu.Lock()
		s.conns[conn] = struct{}{}
		s.mu.Unlock()
		wg.Add(1)
		go s.connHandler(wg, conn)
	}
}

func (s *Server) stopListener() error {
	s.shutdown = true

	return s.sock.Close()
}

// Stop shuts listener and ALL the connections down
func (s *Server) Stop() error {
	if err := s.stopListener(); err != nil {
		return err
	}

	s.mu.Lock()

	for conn := range s.conns {
		_ = conn.Close()
	}

	s.mu.Unlock()

	return nil
}

// GracefulShutdown stops a listener, but leaving all the connections free to end their
// lives peacefully
func (s *Server) GracefulShutdown() error {
	return s.stopListener()
}

func (s *Server) connHandler(wg *sync.WaitGroup, conn net.Conn) {
	s.onConn(conn)
	wg.Done()
	s.mu.Lock()
	delete(s.conns, conn)
	s.mu.Unlock()
}
