package indigo

import (
	"errors"
	"net"

	"github.com/fakefloordiv/indigo/http"
	"github.com/fakefloordiv/indigo/internal/pool"
	httpserver "github.com/fakefloordiv/indigo/internal/server/http"
	"github.com/fakefloordiv/indigo/internal/server/tcp"

	"github.com/fakefloordiv/indigo/http/encodings"
	"github.com/fakefloordiv/indigo/http/headers"
	"github.com/fakefloordiv/indigo/http/url"
	"github.com/fakefloordiv/indigo/internal/alloc"
	"github.com/fakefloordiv/indigo/internal/parser/http1"
	"github.com/fakefloordiv/indigo/internal/render"
	"github.com/fakefloordiv/indigo/router"
	"github.com/fakefloordiv/indigo/settings"
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

	codings        encodings.Decoders
	defaultHeaders map[string][]string

	shutdown chan struct{}
}

// NewApp returns a new application object with initialized shutdown chan
func NewApp(addr string) *Application {
	return &Application{
		addr:           addr,
		codings:        encodings.NewContentDecoders(),
		defaultHeaders: defaultHeaders,
		shutdown:       make(chan struct{}, 1),
	}
}

// AddContentDecoder simply adds a new content decoder
func (a Application) AddContentDecoder(token string, decoder encodings.Decoder) {
	a.codings.Add(token, decoder)
}

// SetDefaultHeaders overrides default headers to a passed ones.
// Doing this, make sure you know what are you doing
func (a *Application) SetDefaultHeaders(headers map[string][]string) {
	a.defaultHeaders = headers
}

// Serve takes a router and someSettings, that must be only 0 or 1 elements
// otherwise, error is returned
// Also, if specified, Accept-Encodings default header's value will be set here
func (a Application) Serve(r router.Router, optionalSettings ...settings.Settings) error {
	if accept, found := a.defaultHeaders["Accept-Encodings"]; found && accept == nil {
		a.defaultHeaders["Accept-Encodings"] = a.codings.Acceptable()
	}

	if onStart, ok := r.(router.OnStarter); ok {
		onStart.OnStart()
	}

	s, err := getSettings(optionalSettings...)
	if err != nil {
		return err
	}

	sock, err := net.Listen("tcp", a.addr)
	if err != nil {
		return err
	}

	return tcp.RunTCPServer(sock, func(conn net.Conn) {
		readBuff := make([]byte, s.TCP.ReadBufferSize)
		client := tcp.NewClient(conn, s.TCP.ReadTimeout, readBuff)

		keyAllocator := alloc.NewAllocator(
			s.Headers.MaxKeyLength*s.Headers.Number.Default,
			s.Headers.MaxKeyLength*s.Headers.Number.Maximal,
		)
		valAllocator := alloc.NewAllocator(
			s.Headers.ValueSpace.Default,
			s.Headers.ValueSpace.Maximal,
		)
		objPool := pool.NewObjectPool[[]string](s.Headers.ValuesObjectPoolSize.Maximal)
		query := url.NewQuery(func() map[string][]byte {
			return make(map[string][]byte, s.URL.Query.DefaultMapSize)
		})
		hdrs := headers.NewHeaders(make(map[string][]string, s.Headers.Number.Default))
		response := http.NewResponse()
		bodyReader := http1.NewBodyReader(client, s.Body)
		request := http.NewRequest(hdrs, query, response, conn, bodyReader)

		startLineBuff := make([]byte, s.URL.MaxLength)
		httpParser := http1.NewHTTPRequestsParser(
			request, keyAllocator, valAllocator, objPool, startLineBuff, s.Headers,
		)

		respBuff := make([]byte, 0, s.HTTP.ResponseBuffSize)
		renderer := render.NewRenderer(respBuff, nil, a.defaultHeaders)

		httpServer := httpserver.NewHTTPServer(r)
		httpServer.Run(client, request, bodyReader, renderer, httpParser)
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

func getSettings(s ...settings.Settings) (settings.Settings, error) {
	switch len(s) {
	case 0:
		return settings.Default(), nil
	case 1:
		return settings.Fill(s[0]), nil
	default:
		return settings.Settings{}, errors.New("too many settings (none or single struct is expected)")
	}
}
