package indigo

import (
	"errors"
	"indigo/http/headers"
	"indigo/http/parser/http1"
	"indigo/http/server"
	"indigo/router"
	settings2 "indigo/settings"
	"indigo/types"
	"net"
)

type Application struct {
	addr string

	shutdown chan bool
}

func NewApp(addr string) Application {
	return Application{
		addr:     addr,
		shutdown: make(chan bool),
	}
}

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
		requestHeaders := make(headers.Headers, settings.Headers.Number.Default)
		headersManager := headers.NewManager(requestHeaders, settings.Headers)
		request, gateway := types.NewRequest(&headersManager)

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
