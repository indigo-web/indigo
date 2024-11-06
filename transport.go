package indigo

import (
	"crypto/tls"
	"fmt"
	"github.com/indigo-web/indigo/config"
	"github.com/indigo-web/indigo/http/crypt"
	"github.com/indigo-web/indigo/http/serve"
	"github.com/indigo-web/indigo/router"
	"github.com/indigo-web/indigo/transport"
	"golang.org/x/crypto/acme/autocert"
	"net"
)

type Transport struct {
	addr          string // must be left intact. Used by App entity only
	inner         transport.Transport
	spawnCallback func(cfg *config.Config, r router.Router) func(net.Conn)
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

func TLS(certs ...tls.Certificate) Transport {
	if len(certs) == 0 {
		panic("need at least one certificate")
	}

	return newTLSTransport(&tls.Config{Certificates: certs})
}

// Autocert tries to automatically issue a certificate for the given domains.
// If operation succeeds, those will be (hopefully) saved into the default cache
// directory, which depends on the OS. If you want to specify the cache directory,
// use AutocertFromCache instead.
func Autocert(domains ...string) Transport {
	return AutocertFromCache(tlsCacheDir(), domains...)
}

// AutocertFromCache tries to automatically issue a certificate for the given domains.
// If the operation succeeds, those will be (hopefully) saved into the provided cache
// directory. It's recommended to use Autocert if there are no explicit needs to set
// custom cache directory.
func AutocertFromCache(cache string, domains ...string) Transport {
	m := &autocert.Manager{
		Prompt: autocert.AcceptTOS,
		Cache:  autocert.DirCache(cache),
	}

	if len(domains) > 0 {
		m.HostPolicy = autocert.HostWhitelist(domains...)
	}

	return newTLSTransport(&tls.Config{GetCertificate: m.GetCertificate})
}

func Cert(cert, key string) tls.Certificate {
	c, err := tls.LoadX509KeyPair(cert, key)
	if err != nil {
		panic(fmt.Errorf("could not load TLS certificate: %s", err))
	}

	return c
}

func newTLSTransport(cfg *tls.Config) Transport {
	return Transport{
		inner: transport.NewTLS(cfg),
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
