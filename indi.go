package indigo

import (
	"github.com/indigo-web/indigo/http/encryption"
	"github.com/indigo-web/indigo/http/serve"
	"github.com/indigo-web/indigo/http/status"
	"github.com/indigo-web/indigo/internal/address"
	"github.com/indigo-web/indigo/internal/tcp"
	"github.com/indigo-web/indigo/router/inbuilt"
	"log"
	"net"
	"sync/atomic"

	"github.com/indigo-web/indigo/config"
	"github.com/indigo-web/indigo/router"
)

// App is just a struct with addr and shutdown channel that is currently
// not used. Planning to replace it with context.WithCancel()
type App struct {
	cfg     config.Config
	hooks   hooks
	sources []source
	errCh   chan error
}

// New returns a new App instance.
func New(addr string) *App {
	app := &App{
		cfg:   config.Default(),
		errCh: make(chan error),
	}

	return app.Listen(addr)
}

// Tune replaces default config.
func (a *App) Tune(cfg config.Config) *App {
	a.cfg = config.Fill(cfg)
	return a
}

// OnStart calls the callback at the moment, when all the servers are started. However,
// it isn't strongly guaranteed that they'll be able to accept new connections immediately
func (a *App) OnStart(cb func()) *App {
	a.hooks.OnStart = cb
	return a
}

// OnBind calls the passed callback for every address, that was bound without any errors
func (a *App) OnBind(cb func(addr string)) *App {
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
func (a *App) Listen(addr string, optListener ...Listener) *App {
	addr = address.Normalize(addr)
	listener := optional(optListener, Listener{})
	if listener.Listen == nil {
		listener.Listen = net.Listen
	}

	a.sources = append(a.sources, source{
		Addr:     addr,
		Listener: listener,
	})
	return a
}

func (a *App) HTTPS(addr string, cert, key string) *App {
	return a.Listen(addr, Listener{
		Listen:  tlsListener(cert, key),
		Handler: a.newHTTPHandler(encryption.TLS),
	})
}

// AutoHTTPS enables HTTPS-mode using autocert or generates self-signed certificates if using
// local host
func (a *App) AutoHTTPS(addr string, domains ...string) *App {
	addr = address.Normalize(addr)

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

	return a.Listen(addr, Listener{
		Listen:  autoTLSListener(domains...),
		Handler: a.newHTTPHandler(encryption.TLS),
	})
}

// Serve starts the web-application. If nil is passed instead of a router, empty inbuilt will
// be used.
func (a *App) Serve(r router.Fabric) error {
	if r == nil {
		r = inbuilt.New()
	}

	servers, err := a.bind(r.Initialize())
	if err != nil {
		return err
	}

	return a.run(servers)
}

func (a *App) bind(r router.Router) ([]*tcp.Server, error) {
	servers := make([]*tcp.Server, 0, len(a.sources))

	for _, src := range a.sources {
		listener, err := src.Listen("tcp", src.Addr)
		if err != nil {
			return nil, err
		}

		if src.Handler == nil {
			src.Handler = a.newHTTPHandler(encryption.Plain)
		}

		servers = append(servers, tcp.NewServer(listener, func(conn net.Conn) {
			src.Handler(conn, r)
		}))
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

func (a *App) newHTTPHandler(enc encryption.Token) listenerHandler {
	return func(conn net.Conn, r router.Router) {
		serve.HTTP1(a.cfg, conn, enc, r)
	}
}

type hooks struct {
	OnStart         func()
	OnListenerStart func(addr string)
	OnStop          func()
}

func callIfNotNil(f func()) {
	if f != nil {
		f()
	}
}

type (
	// listenerFactory is net.Listen-compatible on purpose
	listenerFactory func(network, addr string) (net.Listener, error)
	listenerHandler func(conn net.Conn, r router.Router)
)

type Listener struct {
	Listen  listenerFactory
	Handler listenerHandler
}

type source struct {
	Addr string
	Listener
}

func optional[T any](optionals []T, otherwise T) T {
	if len(optionals) == 0 {
		return otherwise
	}

	return optionals[0]
}
