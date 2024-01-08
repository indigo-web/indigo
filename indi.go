package indigo

import (
	"crypto/tls"
	"fmt"
	"github.com/indigo-web/indigo/http/encryption"
	"github.com/indigo-web/indigo/http/status"
	"github.com/indigo-web/indigo/internal/address"
	"github.com/indigo-web/indigo/router/inbuilt"
	"log"
	"net"
	"sync/atomic"

	"github.com/indigo-web/indigo/internal/server/http"
	"github.com/indigo-web/indigo/internal/server/tcp"
	"github.com/indigo-web/indigo/router"
	"github.com/indigo-web/indigo/settings"
)

type ListenerConstructor func(network, addr string) (net.Listener, error)

// App is just a struct with addr and shutdown channel that is currently
// not used. Planning to replace it with context.WithCancel()
type App struct {
	addr      address.Address
	hooks     hooks
	listeners []Listener
	settings  settings.Settings
	errCh     chan error
}

// New returns a new App instance.
func New(addr string) *App {
	appAddr, err := address.Parse(addr)
	if err != nil {
		panic(fmt.Errorf("indigo: listen: bad addr: %v", err))
	}

	return &App{
		addr:     appAddr,
		settings: settings.Default(),
		errCh:    make(chan error),
	}
}

// Tune replaces default settings.
func (a *App) Tune(s settings.Settings) *App {
	a.settings = settings.Fill(s)
	return a
}

// NotifyOnStart calls the callback at the moment, when all the servers are started. However,
// it isn't strongly guaranteed that they'll be able to accept new connections immediately
func (a *App) NotifyOnStart(cb func()) *App {
	a.hooks.OnStart = cb
	return a
}

// NotifyOnStop calls the callback at the moment, when all the servers are down. It's guaranteed,
// that at the moment as the callback is called, the server isn't able to accept any new connections
// and all the clients are already disconnected
func (a *App) NotifyOnStop(cb func()) *App {
	a.hooks.OnStop = cb
	return a
}

// Listen adds a new listener
func (a *App) Listen(port uint16, enc encryption.Encryption, optionalConstructor ...ListenerConstructor) *App {
	constructor := optional(optionalConstructor, net.Listen)
	if constructor == nil {
		constructor = net.Listen
	}

	a.listeners = append(a.listeners, Listener{
		Port:        port,
		Constructor: constructor,
		Encryption:  enc,
	})

	return a
}

func (a *App) TLS(port uint16, constructor ListenerConstructor) *App {
	a.Listen(port, encryption.TLS, constructor)
	return a
}

func (a *App) HTTPS(port uint16, cert, key string) *App {
	return a.TLS(port, func(network, addr string) (net.Listener, error) {
		certificate, err := tls.LoadX509KeyPair(cert, key)
		if err != nil {
			return nil, err
		}

		return tls.Listen(network, addr, &tls.Config{
			Certificates: []tls.Certificate{certificate},
		})
	})
}

// AutoHTTPS enables HTTPS-mode using autocert or generates self-signed certificates if using
// local host
func (a *App) AutoHTTPS(port uint16, domains ...string) *App {
	if a.addr.IsLocalhost() {
		cert, key, err := generateSelfSignedCert()
		if err != nil {
			log.Printf(
				"WARNING: (*App).AutoHTTPS(...): can't generate self-signed certificate: %v. Disabling TLS",
			)

			return a
		}

		return a.HTTPS(port, cert, key)
	}

	return a.TLS(port, autoTLSListener(domains...))
}

// Serve starts the web-application. If nil is passed instead of a router, empty inbuilt will
// be used.
func (a *App) Serve(r router.Router) error {
	if r == nil {
		r = inbuilt.New()
	}

	if err := r.OnStart(); err != nil {
		return err
	}

	a.Listen(a.addr.Port, encryption.Plain, net.Listen)
	servers, err := a.getServers(a.addr, r)
	if err != nil {
		return err
	}

	return a.run(servers)
}

func (a *App) getServers(addr address.Address, r router.Router) ([]*tcp.Server, error) {
	servers := make([]*tcp.Server, len(a.listeners))

	for i, listener := range a.listeners {
		sock, err := listener.Constructor("tcp", addr.SetPort(listener.Port).String())
		if err != nil {
			return nil, err
		}

		servers[i] = tcp.NewServer(sock, a.newTCPCallback(a.settings, r, listener.Encryption))
	}

	return servers, nil
}

func (a *App) run(servers []*tcp.Server) error {
	var failSilently atomic.Bool

	for _, server := range servers {
		go func(server *tcp.Server) {
			err := server.Start()

			if failSilently.Swap(true) {
				return
			}

			a.errCh <- err
		}(server)
	}

	callIfNotNil(a.hooks.OnStart)
	err := <-a.errCh
	switch err {
	case status.ErrGracefulShutdown:
		tcp.PauseAll(servers)
	default:
		// so basically, any error here (including nil) will stop all the servers. However,
		// in order to be intuitive the best choice is to send status.ErrShutdown
		tcp.StopAll(servers)
	}

	callIfNotNil(a.hooks.OnStop)

	return err
}

// GracefulStop stops accepting new connections and waits until all the already connected clients
// disconnects
func (a *App) GracefulStop() {
	a.errCh <- status.ErrGracefulShutdown
}

// Stop stops the whole application immediately.
func (a *App) Stop() {
	a.errCh <- status.ErrShutdown
}

func (a *App) newTCPCallback(s settings.Settings, r router.Router, enc encryption.Encryption) tcp.OnConn {
	return func(conn net.Conn) {
		client := newClient(a.settings.TCP, conn)
		body := newBody(client, s.Body)
		request := newRequest(s, conn, body)
		request.Env.Encryption = enc
		trans := newTransport(s, request)

		httpServer := http.NewServer(r)
		httpServer.Run(client, request, trans)
	}
}

type hooks struct {
	OnStart, OnStop func()
}

func callIfNotNil(f func()) {
	if f != nil {
		f()
	}
}

type Listener struct {
	Port        uint16
	Constructor ListenerConstructor
	Encryption  encryption.Encryption
}

func optional[T any](optionals []T, otherwise T) T {
	if len(optionals) == 0 {
		return otherwise
	}

	return optionals[0]
}
