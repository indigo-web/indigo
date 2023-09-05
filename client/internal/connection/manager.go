package connection

import (
	"github.com/indigo-web/utils/pool"
	"net"
)

// Manager is basically a bit smarter connection pool. It receives a request,
// and tries to find a free connection, that is able to send it.
type Manager struct {
	pool pool.ObjectPool[net.Conn]
}
