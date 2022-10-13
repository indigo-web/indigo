package indigo

import (
	"errors"
	"net"
	"sync"

	server2 "github.com/fakefloordiv/indigo/internal/server"

	"github.com/fakefloordiv/indigo/http/encodings"
	"github.com/fakefloordiv/indigo/http/headers"
	"github.com/fakefloordiv/indigo/http/url"
	"github.com/fakefloordiv/indigo/internal/alloc"
	"github.com/fakefloordiv/indigo/internal/parser/http1"
	"github.com/fakefloordiv/indigo/internal/render"
	"github.com/fakefloordiv/indigo/router"
	settings2 "github.com/fakefloordiv/indigo/settings"
	"github.com/fakefloordiv/indigo/types"
)

const (
	// not specifying version to avoid problems with vulnerable versions
	// Also not specifying, because otherwise I should add an option to
	// disable such a behaviour. I am too lazy for that. Preferring not
	// specifying version at all
	defaultServer = "indigo"

	// actually, we don't know what content type of body user responds
	// with, so due to rfc2068 7.2.1 it is supposed to be
	// application/octet-stream, but we know that it is usually text/html,
	// isn't it?
	defaultContentType = "text/html"
)

var defaultHeaders = map[string][]string{
	"Server":       {defaultServer},
	"Content-Type": {defaultContentType},
	// nil here means that value will be set later, when server will be initializing
	"Accept-Encodings": nil,
}

// Application is just a struct with addr and shutdown channel that is currently
// not used. Planning to replace it with context.WithCancel()
type Application struct {
	addr string

	codings        encodings.ContentEncodings
	defaultHeaders map[string][]string

	shutdown chan struct{}
}

// NewApp returns a new application object with initialized shutdown chan
func NewApp(addr string) *Application {
	return &Application{
		addr:           addr,
		codings:        encodings.NewContentEncodings(),
		defaultHeaders: defaultHeaders,
		shutdown:       make(chan struct{}, 1),
	}
}

// AddContentDecoder simply adds a new content decoder
func (a Application) AddContentDecoder(token string, decoder encodings.Decoder) {
	a.codings.AddDecoder(token, decoder)
}

// SetDefaultHeaders overrides default headers to a passed ones.
// Doing this, make sure you know what are you doing
func (a *Application) SetDefaultHeaders(headers map[string][]string) {
	a.defaultHeaders = headers
}

// Serve takes a router and someSettings, that must be only 0 or 1 elements
// otherwise, error is returned
// Also, if specified, Accept-Encodings default header's value will be set here
func (a Application) Serve(r router.Router, someSettings ...settings2.Settings) error {
	if accept, found := a.defaultHeaders["Accept-Encodings"]; found && accept == nil {
		a.defaultHeaders["Accept-Encodings"] = a.codings.Acceptable()
	}

	if onStart, ok := r.(router.OnStart); ok {
		onStart.OnStart()
	}

	settings, err := getSettings(someSettings...)
	if err != nil {
		return err
	}

	sock, err := net.Listen("tcp", a.addr)
	if err != nil {
		return err
	}

	return server2.StartTCPServer(sock, func(wg *sync.WaitGroup, conn net.Conn) {
		keyAllocator := alloc.NewAllocator(
			int(settings.Headers.KeyLength.Maximal)*int(settings.Headers.Number.Default),
			int(settings.Headers.KeyLength.Maximal)*int(settings.Headers.Number.Maximal),
		)
		valAllocator := alloc.NewAllocator(
			int(settings.Headers.ValueSpace.Default),
			int(settings.Headers.ValueSpace.Maximal),
		)
		query := url.NewQuery(func() map[string][]byte {
			return make(map[string][]byte, settings.URL.Query.Number.Default)
		})
		hdrs := headers.NewHeaders(make(map[string][]string, settings.Headers.Number.Default))
		request, gateway := types.NewRequest(hdrs, query, conn.RemoteAddr())

		startLineBuff := make([]byte, 0, settings.URL.Length.Default)

		httpParser := http1.NewHTTPRequestsParser(
			request, gateway, keyAllocator, valAllocator, startLineBuff, settings, a.codings,
		)

		respBuff := make([]byte, 0, settings.ResponseBuff.Default)
		renderer := render.NewRenderer(respBuff, a.defaultHeaders)

		httpServer := server2.NewHTTPServer(request, r, httpParser, conn, renderer)
		go httpServer.Run()

		readBuff := make([]byte, settings.TCPServer.Read.Default)
		server2.DefaultConnHandler(
			wg, conn, settings.TCPServer.IDLEConnLifetime, httpServer.OnData, readBuff,
		)
	}, a.shutdown)
}

// Shutdown gracefully shutting down the server. It is not blocking,
// server being shut down right after calling this method is not
// guaranteed, because tcp server will wait for the next connection,
// and only then he'll be able to receive a shutdown notify. Moreover,
// tcp server will wait until all the existing connections will be
// closed
func (a Application) Shutdown() {
	a.shutdown <- struct{}{}
}

// Wait waits for tcp server to shut down
func (a Application) Wait() {
	<-a.shutdown
}

func getSettings(settings ...settings2.Settings) (settings2.Settings, error) {
	switch len(settings) {
	case 0:
		return settings2.Default(), nil
	case 1:
		return settings2.Fill(settings[0]), nil
	default:
		return settings2.Settings{}, errors.New("too many settings (none or single struct is expected)")
	}
}
