package tcp

import (
	"net"
	"sync"
	"sync/atomic"
)

type OnConn func(net.Conn)

type Server struct {
	wg       *sync.WaitGroup
	sock     net.Listener
	onConn   OnConn
	shutdown atomic.Bool
}

func NewServer(sock net.Listener, onConn OnConn) *Server {
	return &Server{
		wg:     new(sync.WaitGroup),
		sock:   sock,
		onConn: onConn,
	}
}

// Start runs the accept-loop until an error during accepting the connection happens
// or graceful shutdown invokes
func (s *Server) Start() error {
	for !s.shutdown.Load() {
		conn, err := s.sock.Accept()
		if err != nil {
			return err
		}

		s.wg.Add(1)
		go s.connHandler(conn)
	}

	return nil
}

// Stop closes the server socket, however all the clients won't be notified explicitly
// until they try to send us anything
func (s *Server) Stop() error {
	return s.sock.Close()
}

// Pause stops listening to new connections, however doesn't close the socket.
//
// NOTE: it doesn't stop the accept-loop immediately, however it waits until the next client connects
func (s *Server) Pause() {
	s.shutdown.Store(true)
}

// Wait blocks the caller until all the connections are closed
func (s *Server) Wait() {
	s.wg.Wait()
}

func (s *Server) connHandler(conn net.Conn) {
	s.onConn(conn)
	s.wg.Done()
}

func StopAll(servers []*Server) {
	for _, server := range servers {
		_ = server.Stop()
	}

	WaitAll(servers)
}

func PauseAll(servers []*Server) {
	for _, server := range servers {
		server.Pause()
	}

	WaitAll(servers)
}

func WaitAll(servers []*Server) {
	for _, server := range servers {
		server.Wait()
	}
}
