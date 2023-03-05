package indigo

import (
	"errors"
	"net"

	"github.com/indigo-web/indigo/http"
	"github.com/indigo-web/indigo/internal/mapconv"
	"github.com/indigo-web/indigo/internal/pool"
	httpserver "github.com/indigo-web/indigo/internal/server/http"
	"github.com/indigo-web/indigo/internal/server/tcp"

	"github.com/indigo-web/indigo/http/encodings"
	"github.com/indigo-web/indigo/http/headers"
	"github.com/indigo-web/indigo/http/query"
	"github.com/indigo-web/indigo/internal/alloc"
	"github.com/indigo-web/indigo/internal/parser/http1"
	"github.com/indigo-web/indigo/internal/render"
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
	addr string

	decoders       encodings.Decoders
	defaultHeaders map[string][]string

	shutdown chan struct{}
}

// NewApp returns a new application object with initialized shutdown chan
func NewApp(addr string) *Application {
	return &Application{
		addr:           addr,
		decoders:       encodings.NewContentDecoders(),
		defaultHeaders: mapconv.Copy(DefaultHeaders),
		shutdown:       make(chan struct{}, 1),
	}
}

// AddContentDecoder simply adds a new content decoder
func (a *Application) AddContentDecoder(token string, decoder encodings.Decoder) {
	a.decoders.Add(token, decoder)
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
		a.defaultHeaders["Accept-Encodings"] = a.decoders.Acceptable()
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
		objPool := pool.NewObjectPool[[]string](s.Headers.MaxValuesObjectPoolSize)
		q := query.NewQuery(func() map[string][]byte {
			return make(map[string][]byte, s.URL.Query.DefaultMapSize)
		})
		hdrs := headers.NewHeaders(make(map[string][]string, s.Headers.Number.Default))
		response := http.NewResponse()
		bodyReader := http1.NewBodyReader(client, s.Body)
		request := http.NewRequest(hdrs, q, response, conn, bodyReader)

		startLineBuff := make([]byte, s.URL.MaxLength)
		httpParser := http1.NewHTTPRequestsParser(
			request, keyAllocator, valAllocator, objPool, startLineBuff, s.Headers,
		)

		respBuff := make([]byte, 0, s.HTTP.ResponseBuffSize)
		renderer := render.NewEngine(respBuff, nil, a.defaultHeaders)

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
func (a *Application) Shutdown() {
	a.shutdown <- struct{}{}
}

// Wait waits for tcp server to shut down
func (a *Application) Wait() {
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
