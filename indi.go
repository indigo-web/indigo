package indigo

import (
	"errors"
	"github.com/fakefloordiv/indigo/http/encodings"
	"net"
	"sync"

	"github.com/fakefloordiv/indigo/http/headers"
	"github.com/fakefloordiv/indigo/http/parser/http1"
	"github.com/fakefloordiv/indigo/http/server"
	"github.com/fakefloordiv/indigo/http/url"
	"github.com/fakefloordiv/indigo/router"
	settings2 "github.com/fakefloordiv/indigo/settings"
	"github.com/fakefloordiv/indigo/types"
)

// Application is just a struct with addr and shutdown channel that is currently
// not used. Planning to replace it with context.WithCancel()
type Application struct {
	addr string

	shutdown chan bool
}

// NewApp returns a new application object with initialized shutdown chan
func NewApp(addr string) Application {
	return Application{
		addr:     addr,
		shutdown: make(chan bool),
	}
}

// Serve takes a router and someSettings, that must be only 0 or 1 elements
// otherwise, error is returned
func (a Application) Serve(r router.Router, someSettings ...settings2.Settings) error {
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

		var codings encodings.ContentEncodings

		if getEncodings, ok := r.(router.GetContentEncodings); ok {
			codings = getEncodings.GetContentEncodings()
		} else {
			codings = encodings.NewContentEncodings()
		}

		httpParser := http1.NewHTTPRequestsParser(
			request, gateway, startLineBuff, headerBuff, settings, &headersManager, codings,
		)

		httpServer := server.NewHTTPServer(request, func(b []byte) (err error) {
			_, err = conn.Write(b)
			return err
		}, r, httpParser, conn)
		go httpServer.Run()

		readBuff := make([]byte, settings.TCPServer.Read.Default)
		server.DefaultConnHandler(wg, conn, readBuff, httpServer.OnData)
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
