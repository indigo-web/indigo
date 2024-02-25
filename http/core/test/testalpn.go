package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"github.com/indigo-web/indigo/internal/transport/http2"
	"math/big"
	"os"
	"strconv"
	"time"
)

const certName = "local.crt"
const keyName = "local.key"

func genCert() error {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return err
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
		return err
	}

	certFile, err := os.Create(certName)
	if err != nil {
		return err
	}
	defer certFile.Close()

	err = pem.Encode(certFile, &pem.Block{Type: "CERTIFICATE", Bytes: certDER})
	if err != nil {
		return err
	}

	keyFile, err := os.Create(keyName)
	if err != nil {
		return err
	}
	defer keyFile.Close()

	privBytes, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		return err
	}

	return pem.Encode(keyFile, &pem.Block{Type: "PRIVATE KEY", Bytes: privBytes})
}

func runTLSServer() error {
	if err := genCert(); err != nil {
		return fmt.Errorf("cannot generate self-signed certificate: %s", err)
	}

	cert, err := tls.LoadX509KeyPair(certName, keyName)
	if err != nil {
		return err
	}

	cfg := &tls.Config{
		Rand:       rand.Reader,
		ServerName: "localhost",
		NextProtos: []string{
			"h2", "http/1.1",
		},
		Certificates:       []tls.Certificate{cert},
		InsecureSkipVerify: true,
	}

	listener, err := tls.Listen("tcp", "localhost:4445", cfg)
	if err != nil {
		return err
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			return err
		}
		go handleTLSConn(conn.(*tls.Conn))
	}
}

func handleTLSConn(conn *tls.Conn) {
	defer conn.Close()

	buff := make([]byte, 2048)
	parser := http2.NewParser()

	for {
		n, err := conn.Read(buff)
		if err != nil {
			fmt.Println(conn.RemoteAddr(), "error:", err)
			return
		}

		fmt.Println(conn.RemoteAddr(), "data:", strconv.Quote(string(buff[:n])))
		state, extra, err := parser.Parse(buff[:n])
		fmt.Println(conn.RemoteAddr(), "parser:", state, extra, err)
		if err != nil {
			return
		}
	}

	//fmt.Println(strconv.Quote(string(buff[:n])))
	//
	//fmt.Printf("%s: err=\"%s\" proto=\"%s\"\n", conn.RemoteAddr().String(), err, conn.ConnectionState().NegotiatedProtocol)
	//conn.Close()
}

func main() {
	fmt.Println(runTLSServer())
}
