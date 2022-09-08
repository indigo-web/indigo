package indigo

import (
	"errors"
	"net"
	"sync"

	"github.com/fakefloordiv/indigo/http/render"

	"github.com/fakefloordiv/indigo/http/encodings"
	"github.com/fakefloordiv/indigo/http/headers"
	"github.com/fakefloordiv/indigo/http/parser/http1"
	"github.com/fakefloordiv/indigo/http/server"
	"github.com/fakefloordiv/indigo/http/url"
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

	// Automatically specify connection as keep-alive. Maybe it is not
	// a compulsory move from server, but still
	defaultConnection = "keep-alive"

	// actually, we don't know what content type of body user responds
	// with, so due to rfc2068 7.2.1 it is supposed to be
	// application/octet-stream, but we know that it is usually text/html,
	// isn't it?
	defaultContentType = "text/html"

	// explicitly set a list of encodings are supported to avoid awkward
	// situations
	// Empty by default because this value is dynamic - depends on which
	// encodings will user add. If it exists, it will be overridden. If
	// it is not existing, this value will not be set
	defaultAcceptEncodings = ""
)

var defaultHeaders = headers.Headers{
	"Server": []headers.Header{
		{Value: defaultServer},
	},
	"Connection": []headers.Header{
		{Value: defaultConnection},
	},
	"Content-Type": []headers.Header{
		{Value: defaultContentType},
	},
	"Accept-Encodings": []headers.Header{
		{Value: defaultAcceptEncodings},
	},
}

// Application is just a struct with addr and shutdown channel that is currently
// not used. Planning to replace it with context.WithCancel()
type Application struct {
	addr string

	codings        encodings.ContentEncodings
	defaultHeaders headers.Headers

	shutdown chan bool
}

// NewApp returns a new application object with initialized shutdown chan
func NewApp(addr string) *Application {
	return &Application{
		addr:           addr,
		codings:        encodings.NewContentEncodings(),
		defaultHeaders: defaultHeaders,
		shutdown:       make(chan bool),
	}
}

// AddContentDecoder simply adds a new content decoder
func (a Application) AddContentDecoder(token string, decoder encodings.Decoder) {
	a.codings.AddDecoder(token, decoder)
}

// SetDefaultHeaders overrides default headers to a passed ones.
// Doing this, make sure you know what are you doing
func (a *Application) SetDefaultHeaders(headers headers.Headers) {
	a.defaultHeaders = headers
}

// Serve takes a router and someSettings, that must be only 0 or 1 elements
// otherwise, error is returned
// Also, if specified, Accept-Encodings default header's value will be set here
func (a Application) Serve(r router.Router, someSettings ...settings2.Settings) error {
	if _, found := a.defaultHeaders["Accept-Encodings"]; found {
		a.defaultHeaders["Accept-Encodings"] = acceptableCodingsHeader(a.codings.Acceptable())
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

	return server.StartTCPServer(sock, func(wg *sync.WaitGroup, conn net.Conn) {
		headersManager := headers.NewManager(settings.Headers)
		query := url.NewQuery(func() map[string][]byte {
			return make(map[string][]byte, settings.URL.Query.Number.Default)
		})
		request, gateway := types.NewRequest(&headersManager, query)

		startLineBuff := make([]byte, 0, settings.URL.Length.Default)
		headerBuff := make([]byte, 0, settings.Headers.KeyLength.Default)

		httpParser := http1.NewHTTPRequestsParser(
			request, gateway, startLineBuff, headerBuff, settings, &headersManager, a.codings,
		)

		renderer := render.NewRenderer(nil, a.defaultHeaders)

		httpServer := server.NewHTTPServer(request, r, httpParser, conn, renderer)
		go httpServer.Run()

		readBuff := make([]byte, settings.TCPServer.Read.Default)
		server.DefaultConnHandler(
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
	a.shutdown <- true
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

func acceptableCodingsHeader(acceptable []string) (values []headers.Header) {
	for i := range acceptable {
		values = append(values, headers.Header{
			Value: acceptable[i],
		})
	}

	return values
}
