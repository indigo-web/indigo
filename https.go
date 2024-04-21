package indigo

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"golang.org/x/crypto/acme/autocert"
	"log"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

func homeDir() string {
	if runtime.GOOS == "windows" {
		return os.Getenv("HOMEDRIVE") + os.Getenv("HOMEPATH")
	}
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return "/"
}

func cacheDir() string {
	const base = "golang-autocert"
	switch runtime.GOOS {
	case "darwin":
		return filepath.Join(homeDir(), "Library", "Caches", base)
	case "windows":
		for _, ev := range []string{"APPDATA", "CSIDL_APPDATA", "TEMP", "TMP"} {
			if v := os.Getenv(ev); v != "" {
				return filepath.Join(v, base)
			}
		}
		// Worst case:
		return filepath.Join(homeDir(), base)
	}
	if xdg := os.Getenv("XDG_CACHE_HOME"); xdg != "" {
		return filepath.Join(xdg, base)
	}
	return filepath.Join(homeDir(), ".cache", base)
}

func tlsListener(cert, key string) listenerFactory {
	return func(network, addr string) (net.Listener, error) {
		certificate, err := tls.LoadX509KeyPair(cert, key)
		if err != nil {
			return nil, err
		}

		return tls.Listen(network, addr, &tls.Config{
			Certificates: []tls.Certificate{certificate},
		})
	}
}

func autoTLSListener(domains ...string) listenerFactory {
	return func(network, addr string) (net.Listener, error) {
		m := &autocert.Manager{
			Prompt: autocert.AcceptTOS,
		}

		if len(domains) > 0 {
			m.HostPolicy = autocert.HostWhitelist(domains...)
		}

		cache := cacheDir()
		if err := mkdirIfNotExists(cache); err != nil {
			log.Printf("WARNING: auto HTTPS: not using a cache: %s", err)
		} else {
			m.Cache = autocert.DirCache(cache)
		}

		cfg := &tls.Config{
			GetCertificate: m.GetCertificate,
		}

		return tls.Listen(network, addr, cfg)
	}
}

func generateSelfSignedCert() (cert, key string, err error) {
	var (
		cache        = cacheDir()
		certFilename = filepath.Join(cache, "localhost.crt")
		keyFilename  = filepath.Join(cache, "localhost.key")
	)

	if certExists(certFilename, keyFilename) {
		return certFilename, keyFilename, nil
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

	certFile, err := os.Create(certFilename)
	if err != nil {
		return "", "", err
	}
	defer certFile.Close()

	err = pem.Encode(certFile, &pem.Block{Type: "CERTIFICATE", Bytes: certDER})
	if err != nil {
		return "", "", err
	}

	keyFile, err := os.Create(keyFilename)
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

	return certFilename, keyFilename, nil
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
