package indigo

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"time"

	"github.com/indigo-web/indigo/config"
	"github.com/indigo-web/indigo/http/codec"
	"github.com/indigo-web/indigo/http/serve"
	"github.com/indigo-web/indigo/internal/codecutil"
	"github.com/indigo-web/indigo/router"
	"github.com/indigo-web/indigo/transport"
	"golang.org/x/crypto/acme/autocert"
)

type Transport struct {
	addr          string
	inner         transport.Transport
	spawnCallback func(cfg *config.Config, r router.Router, c []codec.Codec) func(net.Conn)
}

func TCP() Transport {
	return Transport{
		inner: transport.NewTCP(),
		spawnCallback: func(cfg *config.Config, r router.Router, c []codec.Codec) func(net.Conn) {
			acceptString := codecutil.AcceptEncoding(c)

			return func(conn net.Conn) {
				serve.HTTP1(cfg, conn, 0, r, codecutil.NewCache(c, acceptString))
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
// use AutocertWithCache instead.
func Autocert(domains ...string) Transport {
	return AutocertWithCache(tlsCacheDir(), domains...)
}

// AutocertWithCache tries to automatically issue a certificate for the given domains.
// If the operation succeeds, those will be (hopefully) saved into the provided cache
// directory. It's recommended to use Autocert if there are no explicit needs to set
// custom cache directory.
func AutocertWithCache(cache string, domains ...string) Transport {
	m := &autocert.Manager{
		Prompt: autocert.AcceptTOS,
		Cache:  autocert.DirCache(cache),
	}

	if len(domains) > 0 {
		m.HostPolicy = autocert.HostWhitelist(domains...)
	}

	return newTLSTransport(&tls.Config{GetCertificate: m.GetCertificate})
}

// LocalCert issues a self-signed certificate for local TLS-secured connections.
// Please note, that self-signed certificates are failing security checks, so
// browsers and tools (e.g. curl) may refuse to connect without adding security check skip
// flags (in particular, -k or --insecure for curl.)
func LocalCert(cache ...string) tls.Certificate {
	dir := tlsCacheDir()
	if len(cache) > 0 {
		dir = cache[0]
	}

	cert, key, err := generateSelfSignedCert(dir)
	if err != nil {
		panic(fmt.Errorf("cannot issue a local certificate: %s", err))
	}

	return Cert(cert, key)
}

// Cert loads the TLS certificate. Panics if an error happened.
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
		spawnCallback: func(cfg *config.Config, r router.Router, c []codec.Codec) func(net.Conn) {
			acceptString := codecutil.AcceptEncoding(c)

			return func(conn net.Conn) {
				ver := conn.(*tls.Conn).ConnectionState().Version
				serve.HTTP1(cfg, conn, ver, r, codecutil.NewCache(c, acceptString))
			}
		},
	}
}

func generateSelfSignedCert(cache string) (cert, key string, err error) {
	var (
		certfile = filepath.Join(cache, "localhost.crt")
		keyfile  = filepath.Join(cache, "localhost.key")
	)

	if certExists(certfile, keyfile) {
		return certfile, keyfile, nil
	}

	if err := mkdirIfNotExists(cache); err != nil {
		return "", "", err
	}

	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return "", "", err
	}

	notBefore := time.Now()
	notAfter := notBefore.Add(10 * 365 * 24 * time.Hour) // 10 years validity

	template := x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{Organization: []string{"Localhost"}},
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		return "", "", err
	}

	certFile, err := os.Create(certfile)
	if err != nil {
		return "", "", err
	}
	defer certFile.Close()

	err = pem.Encode(certFile, &pem.Block{Type: "CERTIFICATE", Bytes: certDER})
	if err != nil {
		return "", "", err
	}

	keyFile, err := os.Create(keyfile)
	if err != nil {
		return "", "", err
	}
	defer keyFile.Close()

	privBytes, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		return "", "", err
	}

	err = pem.Encode(keyFile, &pem.Block{Type: "PRIVATE KEY", Bytes: privBytes})
	if err != nil {
		return "", "", err
	}

	return certfile, keyfile, nil
}

func mkdirIfNotExists(dir string) error {
	if stat, err := os.Stat(dir); err == nil && stat.IsDir() {
		return nil
	}

	return os.MkdirAll(dir, 0700)
}

func certExists(cert, key string) bool {
	return fileExists(cert) && fileExists(key)
}

func fileExists(filename string) bool {
	stat, err := os.Stat(filename)

	return err == nil && !stat.IsDir()
}
