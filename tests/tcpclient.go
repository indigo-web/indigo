package tests

import (
	"fmt"
	"net"
)

func idleTCPConn(addr string, port uint16) error {
	sock, err := net.Dial("tcp", fmt.Sprintf("%s:%d", addr, port))
	if err != nil {
		return err
	}

	buff := make([]byte, 15)

	for {
		_, err = sock.Read(buff)
		if err != nil {
			return err
		}
	}
}
