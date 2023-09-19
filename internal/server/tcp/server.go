package tcp

import (
	"github.com/indigo-web/indigo/http/status"
	"net"
	"sync"
)

type onConnection func(net.Conn)

type Server struct {
	sock     net.Listener
	conns    map[net.Conn]struct{}
	shutdown bool
}

func NewServer(sock net.Listener) *Server {
	return &Server{
		sock:  sock,
		conns: map[net.Conn]struct{}{},
	}
}

func (t *Server) Start(onConn onConnection) error {
	wg := new(sync.WaitGroup)

	for {
		conn, err := t.sock.Accept()
		if err != nil {
			wg.Wait()

			if t.shutdown {
				return status.ErrShutdown
			}

			return err
		}

		t.conns[conn] = struct{}{}
		wg.Add(1)
		go connHandler(wg, conn, onConn)
	}
}

func (t *Server) stopListener() error {
	t.shutdown = true

	return t.sock.Close()
}

// Stop shuts listener and ALL the connections down
func (t *Server) Stop() error {
	if err := t.stopListener(); err != nil {
		return err
	}

	for conn := range t.conns {
		_ = conn.Close()
	}

	return nil
}

// GracefulShutdown stops a listener, but leaving all the connections free to end their
// lives peacefully
func (t *Server) GracefulShutdown() error {
	return t.stopListener()
}

func connHandler(wg *sync.WaitGroup, conn net.Conn, onConn onConnection) {
	onConn(conn)
	wg.Done()
}
