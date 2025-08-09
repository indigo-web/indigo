package transport

import (
	"crypto/tls"
	"net"
)

type TLS struct {
	TCP
	cfg *tls.Config
}

func NewTLS(cfg *tls.Config) *TLS {
	return &TLS{cfg: cfg}
}

func (t *TLS) Bind(addr string) error {
	tcp, err := bindTCP(addr)
	if err != nil {
		return err
	}

	l := tls.NewListener(tcp, t.cfg)
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
