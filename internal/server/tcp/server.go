package tcp

import (
	"net"
	"sync"

	"github.com/indigo-web/indigo/v2/http/status"
)

type onConnection func(net.Conn)

// RunTCPServer is an accept-loop with a support of graceful shutdown.
// Graceful shutdown can be initialized by sending a notifier into the sd chan.
// In this case accept-loop still will wait for a connection, but after it'll be
// accepted RunTCPServer will wait for all the connections to be closed, returning
// status.ErrShutdown error in a result
func RunTCPServer(sock net.Listener, onConn onConnection, sd chan struct{}) error {
	wg := new(sync.WaitGroup)

	for {
		select {
		case <-sd:
			wg.Wait()
			sd <- struct{}{}

			return status.ErrShutdown
		default:
			conn, err := sock.Accept()
			if err != nil {
				return err
			}

			wg.Add(1)
			go connHandler(wg, conn, onConn)
		}
	}
}

func connHandler(wg *sync.WaitGroup, conn net.Conn, onConn onConnection) {
	onConn(conn)
	wg.Done()
}
