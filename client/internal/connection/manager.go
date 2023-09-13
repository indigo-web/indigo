package connection

import (
	"github.com/indigo-web/utils/pool"
	"net"
)

type Manager interface {
	Get()
}

// connManager is a basic connections manager, designed for HTTP/1.x mostly
type connManager struct {
	pool pool.ObjectPool[net.Conn]
}
