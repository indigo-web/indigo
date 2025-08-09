package indigo

import (
	"crypto/tls"

	"github.com/indigo-web/indigo/config"
	"github.com/indigo-web/indigo/http/codec"
	"github.com/indigo-web/indigo/internal/strutil"
	"github.com/indigo-web/indigo/router"
	"github.com/indigo-web/indigo/router/inbuilt"
	"github.com/indigo-web/indigo/transport"
)

const Version = "0.17.2"

// App is just a struct with addr and shutdown channel that is currently
// not used. Planning to replace it with context.WithCancel()
type App struct {
	cfg   *config.Config
	hooks struct {
		OnStart func()
		OnBind  func(addr string)
		OnStop  func()
	}
	codecs     []codec.Codec
	transports []Transport
	supervisor transport.Supervisor
}

// New returns a new App instance.
func New(addr string) *App {
	return (&App{
		cfg:        config.Default(),
		supervisor: transport.NewSupervisor(),
	}).Listen(addr, TCP())
}

// Tune replaces default config.
func (a *App) Tune(cfg *config.Config) *App {
	a.cfg = cfg
	return a
}

// OnStart calls the callback at the moment, when all the servers are started. However,
// it isn't strongly guaranteed that they'll be able to accept new connections immediately.
func (a *App) OnStart(cb func()) *App {
	a.hooks.OnStart = cb
	return a
}

// OnBind callback is called every time a listener is ready to accept new connections.
func (a *App) OnBind(cb func(addr string)) *App {
	a.hooks.OnBind = cb
	return a
}

// OnStop calls the callback at the moment, when all the servers are down. It's guaranteed,
// that at the moment as the callback is called, the server isn't able to accept any new connections
// and all the clients are already disconnected.
func (a *App) OnStop(cb func()) *App {
	a.hooks.OnStop = cb
	return a
}

// Codec appends a new codec into the list of supported.
func (a *App) Codec(codec codec.Codec) *App {
	a.codecs = append(a.codecs, codec)
	return a
}

func (a *App) Listen(addr string, ts ...Transport) *App {
	if len(addr) == 0 {
		// empty addr is considered a no-op Bind operation. Main use-case is omitting
		// default TCP binding when, for example, the default TCP listener is preferred
		// to be disabled
		return a
	}

	addr = strutil.NormalizeAddress(addr)

	if len(ts) == 0 {
		return a.Listen(addr, TCP())
	}

	for _, t := range ts {
		t.addr = addr
		a.transports = append(a.transports, t)
	}

	return a
}

// TLS is a shortcut for App.Listen(addr, indigo.TLS(indigo.Cert(cert, key))).
//
// Starts an TLS listener on the provided address using provided 1 or more certificates.
// Zero passed certificates will panic.
func (a *App) TLS(addr string, certs ...tls.Certificate) *App {
	return a.Listen(addr, TLS(certs...))
}

// Serve starts the web-application. If nil is passed instead of a router, empty inbuilt will
// be used.
func (a *App) Serve(r router.Builder) error {
	if r == nil {
		r = inbuilt.New()
	}

	return a.run(r.Build())
}

func (a *App) run(r router.Router) error {
	if a.hooks.OnStart != nil {
		a.hooks.OnStart()
	}

	for _, t := range a.transports {
		if err := a.supervisor.Add(t.addr, t.inner, t.spawnCallback(a.cfg, r, a.codecs)); err != nil {
			return err
		}

		if a.hooks.OnBind != nil {
			a.hooks.OnBind(t.addr)
		}
	}

	err := a.supervisor.Run(a.cfg.NET)
	if a.hooks.OnStop != nil {
		a.hooks.OnStop()
	}

	return err
}

// Stop stops the whole application immediately and waits until it _really_ stops.
func (a *App) Stop() {
	a.supervisor.Stop()
}
