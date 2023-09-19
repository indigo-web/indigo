package indigo

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"

	httpserver "github.com/indigo-web/indigo/internal/server/http"
	"github.com/indigo-web/indigo/internal/server/tcp"
	"github.com/indigo-web/utils/mapconv"

	"github.com/indigo-web/indigo/http/decoder"
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
	ctx              context.Context
	decoders         map[string]decoder.Constructor
	defaultHeaders   map[string][]string
	addr             string
	servers          []*tcp.Server
	gracefulShutdown bool
}

// NewApp returns a new application object with initialized shutdown chan
func NewApp(addr string) *Application {
	return &Application{
		addr:           addr,
		decoders:       map[string]decoder.Constructor{},
		defaultHeaders: mapconv.Copy(DefaultHeaders),
	}
}

// AddContentDecoder simply adds a new content decoder
func (a *Application) AddContentDecoder(token string, decoder decoder.Constructor) {
	a.decoders[token] = decoder
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

func (a *Application) SetContext(ctx context.Context) {
	a.ctx = ctx
}

// Serve takes a router and someSettings, that must be only 0 or 1 elements
// otherwise, error is returned
// Also, if specified, Accept-Encodings default header's value will be set here
func (a *Application) Serve(r router.Router, optionalSettings ...settings.Settings) error {
	if accept, found := a.defaultHeaders["Accept-Encodings"]; found && accept == nil {
		// because of the special treatment of default headers by rendering engine, better to
		// join these values manually. Otherwise, each value will be rendered individually, that
		// still follows the standard, but brings some unnecessary networking overhead
		a.defaultHeaders["Accept-Encodings"] = []string{strings.Join(mapconv.Keys(a.decoders), ",")}
	}

	if err := r.OnStart(); err != nil {
		return err
	}

	s, err := concreteSettings(optionalSettings...)
	if err != nil {
		return err
	}

	if s.TLS.Enable {
		tlsAddr, err := replacePort(a.addr, s.TLS.Port)
		if err != nil {
			return err
		}

		cert, err := tls.LoadX509KeyPair(s.TLS.Cert, s.TLS.Key)
		if err != nil {
			return err
		}

		cfg := &tls.Config{Certificates: []tls.Certificate{cert}}
		listener, err := tls.Listen("tcp", tlsAddr, cfg)
		if err != nil {
			return err
		}

		a.servers = append(a.servers, tcp.NewServer(listener, func(conn net.Conn) {
			client := newClient(s.TCP, conn)
			bodyReader := newBodyReader(client, s.Body, a.decoders)
			request := newRequest(a.ctx, s, conn, bodyReader)
			request.IsTLS = true
			renderer := newRenderer(s.HTTP, a)
			httpParser := newHTTPParser(s, request)

			httpServer := httpserver.NewHTTPServer(r)
			httpServer.Run(client, request, request.Body(), renderer, httpParser)
		}))
	}

	sock, err := net.Listen("tcp", a.addr)
	if err != nil {
		return err
	}

	a.servers = append(a.servers, tcp.NewServer(sock, func(conn net.Conn) {
		client := newClient(s.TCP, conn)
		bodyReader := newBodyReader(client, s.Body, a.decoders)
		request := newRequest(a.ctx, s, conn, bodyReader)
		renderer := newRenderer(s.HTTP, a)
		httpParser := newHTTPParser(s, request)

		httpServer := httpserver.NewHTTPServer(r)
		httpServer.Run(client, request, request.Body(), renderer, httpParser)
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
func concreteSettings(s ...settings.Settings) (settings.Settings, error) {
	switch len(s) {
	case 0:
		return settings.Default(), nil
	case 1:
		return settings.Fill(s[0]), nil
	default:
		return settings.Settings{}, errors.New("too many settings (none or single struct is expected)")
	}
}

func replacePort(addr string, newPort uint16) (newAddr string, err error) {
	host, err := hostFromAddr(addr)
	if err != nil {
		return "", err
	}

	return host + ":" + strconv.Itoa(int(newPort)), nil
}

func hostFromAddr(addr string) (host string, err error) {
	semicolon := strings.IndexByte(addr, ':')
	if semicolon == -1 {
		return "", fmt.Errorf("bad address: %s", addr)
	}

	return addr[:semicolon], nil
}
