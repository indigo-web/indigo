package indigo

import (
	"crypto/tls"
	"fmt"
	"github.com/indigo-web/indigo/http/coding"
	"github.com/indigo-web/indigo/internal/address"
	"net"
	"strings"

	"github.com/indigo-web/indigo/internal/server/http"
	"github.com/indigo-web/indigo/internal/server/tcp"
	"github.com/indigo-web/utils/mapconv"

	"github.com/indigo-web/indigo/router"
	"github.com/indigo-web/indigo/settings"
)

// DefaultHeaders are headers that are going to be sent unless they were overridden by
// user.
//
// WARNING: if you want to edit them, do it using Application.AddDefaultHeader or
// Application.DeleteDefaultHeader instead
var DefaultHeaders = map[string][]string{
	// nil here means that value will be set later, when server will be initializing
	"Accept-Encodings": nil,
}

// Application is just a struct with addr and shutdown channel that is currently
// not used. Planning to replace it with context.WithCancel()
type Application struct {
	codings          []coding.Constructor
	defaultHeaders   map[string][]string
	addr             string
	servers          []*tcp.Server
	gracefulShutdown bool
}

// NewApp returns a new application object with initialized shutdown chan
func NewApp(addr string) *Application {
	return &Application{
		addr:           addr,
		defaultHeaders: mapconv.Copy(DefaultHeaders),
	}
}

// AddCoding adds a new content coding, available for both encoding and decoding
func (a *Application) AddCoding(constructor coding.Constructor) {
	a.codings = append(a.codings, constructor)
}

// SetDefaultHeaders overrides default headers to a passed ones.
// Doing this, make sure you know what are you doing
func (a *Application) SetDefaultHeaders(headers map[string][]string) {
	a.defaultHeaders = headers
}

func (a *Application) AddDefaultHeader(key string, values ...string) {
	a.defaultHeaders[key] = append(a.defaultHeaders[key], values...)
}

func (a *Application) DeleteDefaultHeader(key string) {
	delete(a.defaultHeaders, key)
}

// Serve takes a router and someSettings, that must be only 0 or 1 elements
// otherwise, error is returned
// Also, if specified, Accept-Encodings default header's value will be set here
func (a *Application) Serve(r router.Router, optionalSettings ...settings.Settings) error {
	addr, err := address.Parse(a.addr)
	if err != nil {
		return fmt.Errorf("bad address: %s", err.Error())
	}

	if accept, found := a.defaultHeaders["Accept-Encodings"]; found && accept == nil {
		// because of the special treatment of default headers by rendering engine, better to
		// join these values manually. Otherwise, each value will be rendered individually, that
		// still follows the standard, but brings some unnecessary networking overhead
		acceptTokens := strings.Join(availableEncodings(a.codings...), ",")
		a.defaultHeaders["Accept-Encodings"] = []string{acceptTokens}
	}

	if err := r.OnStart(); err != nil {
		return err
	}

	s := concreteSettings(optionalSettings...)

	if s.TLS.Enable {
		cert, err := tls.LoadX509KeyPair(s.TLS.Cert, s.TLS.Key)
		if err != nil {
			return err
		}

		cfg := &tls.Config{Certificates: []tls.Certificate{cert}}
		listener, err := tls.Listen("tcp", addr.SetPort(s.TLS.Port).String(), cfg)
		if err != nil {
			return err
		}

		a.servers = append(a.servers, tcp.NewServer(listener, func(conn net.Conn) {
			client := newClient(s.TCP, conn)
			bodyReader := newBody(client, s.Body, a.codings)
			request := newRequest(s, conn, bodyReader)
			request.Env.IsTLS = true
			renderer := newRenderer(s.HTTP, a)
			httpParser := newHTTPParser(s, request)

			httpServer := http.NewServer(r)
			httpServer.Run(client, request, renderer, httpParser)
		}))
	}

	sock, err := net.Listen("tcp", addr.String())
	if err != nil {
		return err
	}

	a.servers = append(a.servers, tcp.NewServer(sock, func(conn net.Conn) {
		client := newClient(s.TCP, conn)
		bodyReader := newBody(client, s.Body, a.codings)
		request := newRequest(s, conn, bodyReader)
		renderer := newRenderer(s.HTTP, a)
		httpParser := newHTTPParser(s, request)

		httpServer := http.NewServer(r)
		httpServer.Run(client, request, renderer, httpParser)
	}))

	err = a.startServers()

	if a.gracefulShutdown {
		return nil
	}

	a.stopServers()

	return err
}

func (a *Application) startServers() error {
	// making the channel buffered isn't necessary, but in case something went wrong,
	// no goroutines will leak, as each will write its error into the channel and die
	// peacefully instead of being stuck for forever
	errCh := make(chan error, len(a.servers))

	for _, server := range a.servers {
		go func(server *tcp.Server) {
			errCh <- server.Start()
		}(server)
	}

	return <-errCh
}

// GracefulShutdown stops the application peacefully. It stops a listener, so no more
// clients will be able to connect, but all the already connected will be processed till
// end (till the last one disconnects)
func (a *Application) GracefulShutdown() {
	a.gracefulShutdown = true
	a.gracefulShutdownServers()
}

func (a *Application) Stop() {
	a.stopServers()
}

func (a *Application) stopServers() {
	for _, server := range a.servers {
		_ = server.Stop()
	}
}

func (a *Application) gracefulShutdownServers() {
	for _, server := range a.servers {
		_ = server.GracefulShutdown()
	}
}

// concreteSettings converts optional settings to concrete ones
func concreteSettings(s ...settings.Settings) settings.Settings {
	switch len(s) {
	case 0:
		return settings.Default()
	case 1:
		return settings.Fill(s[0])
	default:
		panic("too many settings. None or single instance is expected")
	}
}

func availableEncodings(codings ...coding.Constructor) []string {
	if len(codings) == 0 {
		return []string{"identity"}
	}

	available := make([]string, 0, len(codings))

	for _, constructor := range codings {
		available = append(available, constructor(nil).Token())
	}

	return available
}
