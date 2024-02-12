package indigo

import (
	"crypto/tls"
	"github.com/indigo-web/indigo/http/encryption"
	"github.com/indigo-web/indigo/http/status"
	"github.com/indigo-web/indigo/internal/address"
	"github.com/indigo-web/indigo/internal/initialize"
	"github.com/indigo-web/indigo/router/inbuilt"
	"log"
	"net"
	"sync/atomic"

	"github.com/indigo-web/indigo/internal/server/http"
	"github.com/indigo-web/indigo/internal/server/tcp"
	"github.com/indigo-web/indigo/router"
	"github.com/indigo-web/indigo/settings"
)

type (
	// ListenerConstructor constructs the listener by itself. It's made net.Listen compatible
	// on a reason
	ListenerConstructor func(network, addr string) (net.Listener, error)
	// MakeListener returns the listener constructor and the encryption token
	MakeListener func() (encryption.Token, ListenerConstructor)
)

// App is just a struct with addr and shutdown channel that is currently
// not used. Planning to replace it with context.WithCancel()
type App struct {
	hooks     hooks
	listeners []Listener
	settings  settings.Settings
	errCh     chan error
}

// New returns a new App instance.
func New(addr string) *App {
	app := &App{
		settings: settings.Default(),
		errCh:    make(chan error),
	}

	return app.Listen(addr)
}

// Tune replaces default settings.
func (a *App) Tune(s settings.Settings) *App {
	a.settings = settings.Fill(s)
	return a
}

// OnStart calls the callback at the moment, when all the servers are started. However,
// it isn't strongly guaranteed that they'll be able to accept new connections immediately
func (a *App) OnStart(cb func()) *App {
	a.hooks.OnStart = cb
	return a
}

func (a *App) OnListenerStart(cb func(Listener)) *App {
	a.hooks.OnListenerStart = cb
	return a
}

// OnStop calls the callback at the moment, when all the servers are down. It's guaranteed,
// that at the moment as the callback is called, the server isn't able to accept any new connections
// and all the clients are already disconnected
func (a *App) OnStop(cb func()) *App {
	a.hooks.OnStop = cb
	return a
}

// Listen adds a new listener
func (a *App) Listen(addr string, optMaker ...MakeListener) *App {
	addr = address.Format(addr)
	constructor := optional(optMaker, makeNetListener)
	if constructor == nil {
		constructor = makeNetListener
	}

	a.listeners = append(a.listeners, newListener(addr, constructor))

	return a
}

func (a *App) TLS(addr string, constructor ListenerConstructor) *App {
	a.Listen(addr, func() (encryption.Token, ListenerConstructor) {
		return encryption.TLS, constructor
	})
	return a
}

func (a *App) HTTPS(addr string, cert, key string) *App {
	return a.TLS(addr, func(network, addr string) (net.Listener, error) {
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
func (a *App) AutoHTTPS(addr string, domains ...string) *App {
	addr = address.Format(addr)

	if address.IsLocalhost(addr) || address.IsIP(addr) {
		cert, key, err := generateSelfSignedCert()
		if err != nil {
			log.Printf(
				"WARNING: AutoHTTPS(...): can't generate self-signed certificate: %s. Disabling TLS",
				err,
			)

			return a
		}

		return a.HTTPS(addr, cert, key)
	}

	return a.TLS(addr, autoTLSListener(domains...))
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

	servers, err := a.getServers(r)
	if err != nil {
		return err
	}

	return a.run(servers)
}

func (a *App) getServers(r router.Router) ([]*tcp.Server, error) {
	servers := make([]*tcp.Server, len(a.listeners))

	for i, listener := range a.listeners {
		enc, l := listener.Make()
		sock, err := l("tcp", listener.Addr)
		if err != nil {
			return nil, err
		}

		servers[i] = tcp.NewServer(sock, a.newTCPCallback(r, enc))

		if cb := a.hooks.OnListenerStart; cb != nil {
			cb(listener)
		}
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
	if err == status.ErrGracefulShutdown {
		// stop listening to new clients and process till the end all the old ones
		tcp.PauseAll(servers)
	}

	tcp.StopAll(servers)
	callIfNotNil(a.hooks.OnStop)

	return err
}

// GracefulStop stops accepting new connections, but keeps serving old ones.
//
// NOTE: the call isn't blocking. So by that, after the method returned, the server
// will be still working
func (a *App) GracefulStop() {
	a.errCh <- status.ErrGracefulShutdown
}

// Stop stops the whole application immediately.
//
// NOTE: the call isn't blocking. So by that, after the method returned, the server
// will still be working
func (a *App) Stop() {
	a.errCh <- status.ErrShutdown
}

func (a *App) newTCPCallback(r router.Router, enc encryption.Token) tcp.OnConn {
	return func(conn net.Conn) {
		client := initialize.NewClient(a.settings.TCP, conn)
		body := initialize.NewBody(client, a.settings.Body)
		request := initialize.NewRequest(a.settings, conn, body)
		request.Env.Encryption = enc
		trans := initialize.NewTransport(a.settings, request)
		httpServer := http.NewServer(r, a.settings.HTTP.OnDisconnect)
		httpServer.Run(client, request, trans)
	}
}

type hooks struct {
	OnStart         func()
	OnListenerStart func(Listener)
	OnStop          func()
}

func callIfNotNil(f func()) {
	if f != nil {
		f()
	}
}

type Listener struct {
	Addr string
	Make MakeListener
}

func newListener(addr string, constructor MakeListener) Listener {
	return Listener{
		Addr: addr,
		Make: constructor,
	}
}

func makeNetListener() (encryption.Token, ListenerConstructor) {
	return encryption.Plain, net.Listen
}

func optional[T any](optionals []T, otherwise T) T {
	if len(optionals) == 0 {
		return otherwise
	}

	return optionals[0]
}
