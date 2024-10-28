package indigo

import (
	"crypto/tls"
	"github.com/indigo-web/indigo/config"
	"github.com/indigo-web/indigo/http/crypt"
	"github.com/indigo-web/indigo/http/serve"
	"github.com/indigo-web/indigo/router"
	"github.com/indigo-web/indigo/transport"
	"net"
)

type Transport struct {
	addr          string // must be left intact. Used by App entity only
	inner         transport.Transport
	spawnCallback func(cfg *config.Config, r router.Router) func(net.Conn)
	error         error
}

func TCP() Transport {
	return Transport{
		inner: transport.NewTCP(),
		spawnCallback: func(cfg *config.Config, r router.Router) func(net.Conn) {
			return func(conn net.Conn) {
				serve.HTTP1(cfg, conn, crypt.Plain, r)
			}
		},
	}
}

func TLS(cert, key string) Transport {
	c, err := tls.LoadX509KeyPair(cert, key)
	if err != nil {
		// if any error occurred, there's no way to report it at this point.
		// Save it in the transport, the App will catch and return it when
		// will bind listeners. Deferred error?
		return Transport{error: err}
	}

	return Transport{
		inner: transport.NewTLS([]tls.Certificate{c}),
		spawnCallback: func(cfg *config.Config, r router.Router) func(net.Conn) {
			return func(conn net.Conn) {
				ver := conn.(*tls.Conn).ConnectionState().Version
				serve.HTTP1(cfg, conn, tlsver2crypttoken(ver), r)
			}
		},
	}
}

func tlsver2crypttoken(ver uint16) crypt.Encryption {
	switch ver {
	case tls.VersionTLS10:
		return crypt.TLSv10
	case tls.VersionTLS11:
		return crypt.TLSv11
	case tls.VersionTLS12:
		return crypt.TLSv12
	case tls.VersionTLS13:
		return crypt.TLSv13
	case tls.VersionSSL30:
		return crypt.SSL
	default:
		return crypt.Unknown
	}
}
