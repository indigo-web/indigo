package server

import (
	"net"
	"sync"

	"github.com/fakefloordiv/indigo/errors"
)

type (
	connHandler func(*sync.WaitGroup, net.Conn)
	dataHandler func([]byte) error
)

// StartTCPServer just starts an accept-loop, starting a goroutine provided as a
// handleConn callback for each connection
// If a value from sd (ShutDown channel) is received (no matter which one), server
// will wait until all the goroutines will die and only then release an execution
// flow
func StartTCPServer(sock net.Listener, handleConn connHandler, sd chan bool) error {
	wg := new(sync.WaitGroup)

	for {
		select {
		case <-sd:
			wg.Wait()
			return errors.ErrShutdown
		default:
			conn, err := sock.Accept()
			if err != nil {
				return err
			}

			wg.Add(1)
			go handleConn(wg, conn)
		}
	}
}

// DefaultConnHandler is a core handler. It takes a buffer for reading provided
// by caller, and starts reading from socket. Guarantees:
// 1) connection will be closed and waitgroup released
//    * in case callback returned errors.ErrHijackConn, connection will not
//      be closed
// 2) on client disconnect, handleData will be called with an empty slice as indicator
//    of disconnect (normally it is not possible)
// Errors occurred while reading socket, will be ignored and user will only know that
// client has disconnected, even if server is guilty
func DefaultConnHandler(wg *sync.WaitGroup, conn net.Conn, buff []byte, handleData dataHandler) {
	defer wg.Done()

	for {
		n, err := conn.Read(buff)
		err2 := handleData(buff[:n])

		if err2 != nil || err != nil || n == 0 {
			if err2 != errors.ErrHijackConn {
				conn.Close()
			}

			return
		}
	}
}
