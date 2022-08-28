package indigo

import (
	"errors"
	"indigo/http/headers"
	"indigo/http/parser/http1"
	"indigo/http/server"
	"indigo/http/url"
	"indigo/router"
	settings2 "indigo/settings"
	"indigo/types"
	"net"
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
func (a Application) Serve(router router.Router, someSettings ...settings2.Settings) error {
	settings, err := getSettings(someSettings...)
	if err != nil {
		return err
	}

	sock, err := net.Listen("tcp", a.addr)
	if err != nil {
		return err
	}

	return server.StartTCPServer(sock, func(conn net.Conn) {
		headersManager := headers.NewManager(settings.Headers)
		query := url.NewQuery(func() map[string][]byte {
			return make(map[string][]byte, settings.URL.Query.Number.Default)
		})
		request, gateway := types.NewRequest(&headersManager, query)

		startLineBuff := make([]byte, 0, settings.URL.Length.Default)
		headerBuff := make([]byte, 0, settings.Headers.KeyLength.Default)

		httpParser := http1.NewHTTPRequestsParser(
			request, gateway, startLineBuff, headerBuff, settings, &headersManager,
		)

		httpServer := server.NewHTTPServer(request, func(b []byte) error {
			_, err := conn.Write(b)
			return err
		}, router, httpParser)

		go httpServer.Run()

		readBuff := make([]byte, settings.TCPServer.Read.Default)
		server.DefaultConnHandler(conn, readBuff, httpServer.OnData)
	})
}

func getSettings(settings ...settings2.Settings) (settings2.Settings, error) {
	switch len(settings) {
	case 0:
		return settings2.Default(), nil
	case 1:
		return settings2.Fill(settings[0]), nil
	default:
		return settings2.Settings{}, errors.New("too much settings")
	}
}
