package server

import (
	"indigo/errors"
	"net"
	"sync"
)

/*
TCP server is a first layer in the web-server, it is responsible for all the low-level activity directly
with sockets. All the received data it provides to
*/

const ReadBytesPerOnce = 2048 // may be even decreased to 1024

type (
	connHandler func(*sync.WaitGroup, net.Conn)
	dataHandler func([]byte) error
)

func StartTCPServer(sock net.Listener, handleConn connHandler, shutdown chan bool) error {
	var wg *sync.WaitGroup

	for {
		conn, err := sock.Accept()

		if err != nil {
			// TODO: high-level api must catch and wait some time before
			//       calling this function again
			return err
		}

		wg.Add(1)
		go handleConn(wg, conn)

		select {
		case <-shutdown:
			wg.Wait()
			shutdown <- true
			return errors.ErrServerShutdown
		default:
		}
	}
}

func DefaultConnHandler(wg *sync.WaitGroup, conn net.Conn, handleData dataHandler) {
	defer conn.Close()
	defer wg.Done()
	buff := make([]byte, ReadBytesPerOnce)

	for {
		n, err := conn.Read(buff)

		if n == 0 || err != nil || handleData(buff[:n]) != nil {
			return
		}
	}
}
