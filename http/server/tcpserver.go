package server

import (
	"net"
)

type (
	connHandler func(net.Conn)
	dataHandler func([]byte) error
)

func StartTCPServer(sock net.Listener, handleConn connHandler) error {
	for {
		conn, err := sock.Accept()

		if err != nil {
			return err
		}

		go handleConn(conn)
	}
}

func DefaultConnHandler(conn net.Conn, buff []byte, handleData dataHandler) {
	defer conn.Close()

	for {
		n, err := conn.Read(buff)
		if handleData(buff[:n]) != nil || n == 0 || err != nil {
			return
		}
	}
}
