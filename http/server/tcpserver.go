package server

import (
	"net"
)

type (
	connHandler func(net.Conn)
	dataHandler func([]byte) error
)

// StartTCPServer just starts an accept-loop, starting a goroutine provided as a
// handleConn callback for each connection
func StartTCPServer(sock net.Listener, handleConn connHandler) error {
	for {
		conn, err := sock.Accept()
		if err != nil {
			return err
		}

		go handleConn(conn)
	}
}

// DefaultConnHandler is a core handler. It takes a buffer for reading provided
// by caller, and starts reading from socket. Guarantees:
// 1) connection will be closed
// 2) on client disconnect, handleData will be called with an empty slice as indicator
//    of disconnect (normally it is not possible)
// Errors occurred while reading socket, will be ignored and user will only know that
// client has disconnected, even if server is guilty
func DefaultConnHandler(conn net.Conn, buff []byte, handleData dataHandler) {
	defer conn.Close()

	for {
		n, err := conn.Read(buff)
		if handleData(buff[:n]) != nil || n == 0 || err != nil {
			return
		}
	}
}
