package testutils

import (
	"fmt"
	"net"
)

func getTCPSock(addr string, port uint16) (net.Listener, error) {
	return net.Listen("tcp", fmt.Sprintf("%s:%d", addr, port))
}
