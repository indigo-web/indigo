package indigo

import (
	"errors"
	"net"
	"strings"

	"github.com/indigo-web/indigo/internal/mapconv"
	httpserver "github.com/indigo-web/indigo/internal/server/http"
	"github.com/indigo-web/indigo/internal/server/tcp"

	"github.com/indigo-web/indigo/http/decode"
	"github.com/indigo-web/indigo/internal/parser/http1"
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
	decoder        *decode.Decoder
	defaultHeaders map[string][]string
	shutdown       chan struct{}
	addr           string
}

// NewApp returns a new application object with initialized shutdown chan
func NewApp(addr string) *Application {
	return &Application{
		addr:           addr,
		decoder:        decode.NewDecoder(),
		defaultHeaders: mapconv.Copy(DefaultHeaders),
		shutdown:       make(chan struct{}, 1),
	}
}

// AddContentDecoder simply adds a new content decoder
func (a *Application) AddContentDecoder(token string, decoder decode.DecoderFactory) {
	a.decoder.Add(token, decoder)
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
	if accept, found := a.defaultHeaders["Accept-Encodings"]; found && accept == nil {
		// because of the special treatment of default headers by rendering engine, better to
		// join these values manually. Otherwise, each value will be rendered individually, that
		// still follows the standard, but brings some unnecessary networking overhead
		a.defaultHeaders["Accept-Encodings"] = []string{strings.Join(a.decoder.Acceptable(), ",")}
	}

	if err := r.OnStart(); err != nil {
		return err
	}

	s, err := concreteSettings(optionalSettings...)
	if err != nil {
		return err
	}

	sock, err := net.Listen("tcp", a.addr)
	if err != nil {
		return err
	}

	return tcp.RunTCPServer(sock, func(conn net.Conn) {
		client := newClient(s.TCP, conn)
		bodyReader := http1.NewBodyReader(client, s.Body)
		request := newRequest(s, conn, bodyReader, a.decoder)
		renderer := newRenderer(s.HTTP, a)
		httpParser := newHTTPParser(s, request)

		httpServer := httpserver.NewHTTPServer(r)
		httpServer.Run(client, request, request.Body(), renderer, httpParser)
	}, a.shutdown)
}

// Shutdown gracefully shutting down the server. It is not blocking,
// server being shut down right after calling this method is not
// guaranteed, because tcp server will wait for the next connection,
// and only then he'll be able to receive a shutdown notify. Moreover,
// tcp server will wait until all the existing connections will be
// closed
func (a *Application) Shutdown() {
	a.shutdown <- struct{}{}
}

// Wait waits for tcp server to shut down
func (a *Application) Wait() {
	<-a.shutdown
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
