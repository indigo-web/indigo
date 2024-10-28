package transport

import (
	"crypto/tls"
	"net"
)

type TLS struct {
	certs []tls.Certificate
	TCP
}

func NewTLS(certs []tls.Certificate) *TLS {
	return &TLS{certs: certs}
}

func (t *TLS) Bind(addr string) error {
	tcp, err := bindTCP(addr)
	if err != nil {
		return err
	}

	l := tls.NewListener(tcp, &tls.Config{
		Certificates: t.certs,
	})
	t.TCP = newTCP(tlsAdapter{tcp, l})

	return nil
}

type tlsAdapter struct {
	*net.TCPListener
	tls net.Listener
}

func (t tlsAdapter) Accept() (net.Conn, error) {
	return t.tls.Accept()
}
