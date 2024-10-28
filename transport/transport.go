package transport

import (
	"github.com/indigo-web/indigo/config"
	"net"
)

type Transport interface {
	Bind(addr string) error
	Listen(cfg config.TCP, cb func(conn net.Conn)) error
	Stop()
	Close()
	Wait()
}
