package tcp

import (
	"github.com/indigo-web/indigo/http/status"
	"net"
	"sync"
)

type onConnection func(net.Conn)

type Server struct {
	sock     net.Listener
	onConn   onConnection
	conns    map[net.Conn]struct{}
	shutdown bool
}

func NewServer(sock net.Listener, onConn onConnection) *Server {
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

		s.conns[conn] = struct{}{}
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

	for conn := range s.conns {
		_ = conn.Close()
	}

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
	delete(s.conns, conn)
}
