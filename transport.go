package indigo

import (
	"crypto/tls"
	"errors"
	"github.com/indigo-web/indigo/config"
	"github.com/indigo-web/indigo/http/crypt"
	"github.com/indigo-web/indigo/http/serve"
	"github.com/indigo-web/indigo/router"
	"github.com/indigo-web/indigo/transport"
	"net"
)

var (
	ErrBadCertificate = errors.New("one or more passed certificates are empty")
	ErrNoCertificates = errors.New("no certificates were passed")
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

	return HTTPS(c)
}

func HTTPS(certs ...tls.Certificate) Transport {
	// simple anti-idiot checks in order to avoid the most obvious mistakes
	switch {
	case len(certs) == 0:
		return Transport{error: ErrNoCertificates}
	case !noEmptyCerts(certs):
		return Transport{error: ErrBadCertificate}
	}

	return Transport{
		inner: transport.NewTLS(certs),
		spawnCallback: func(cfg *config.Config, r router.Router) func(net.Conn) {
			return func(conn net.Conn) {
				ver := conn.(*tls.Conn).ConnectionState().Version
				serve.HTTP1(cfg, conn, tlsver2crypttoken(ver), r)
			}
		},
	}
}

func Cert(cert, key string) tls.Certificate {
	// in case of an error an empty certificate is returned. This will be
	// checked and instantly reported on starting the application
	c, _ := tls.LoadX509KeyPair(cert, key)
	return c
}

func noEmptyCerts(certs []tls.Certificate) bool {
	for _, c := range certs {
		if c.Certificate == nil {
			return false
		}
	}

	return true
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
