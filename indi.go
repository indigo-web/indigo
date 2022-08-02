package indigo

import (
	"fmt"
	"indigo/errors"
	"indigo/http/parser"
	"indigo/http/server"
	"indigo/router"
	"indigo/settings"
	"indigo/types"
	"net"
	"sync"
)

// Application TODO: add domain to allow host multiple
//                   domains in a single web-server instance
type Application struct {
	host string
	port uint16

	shutdown chan bool
}

func NewApp(host string, port uint16) *Application {
	return &Application{
		host: host,
		port: port,
	}
}

func (a Application) Serve(router router.Router, maybeSettings ...settings.Settings) error {
	serverSettings, err := getSettings(maybeSettings)
	if err != nil {
		return err
	}

	address := fmt.Sprintf("%s:%d", a.host, a.port)
	sock, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}
	defer sock.Close()

	return runTCPServer(sock, router, serverSettings, a.shutdown)
}

// Shutdown is a graceful shutdown that waits until all the clients
// will disconnect by their own, without forced disconnecting
// Calling this method blocks until next client will be connected
// (and disconnected instantly)
func (a Application) Shutdown() {
	a.shutdown <- true
	<-a.shutdown
}

func runTCPServer(sock net.Listener, router router.Router,
	serverSettings settings.Settings, shutdown chan bool) error {
	return server.StartTCPServer(sock, func(wg *sync.WaitGroup, conn net.Conn) {
		request, pipe := types.NewRequest(
			// TODO: add all this shit to settings
			make([]byte, 10), make(map[string][]byte), nil,
			serverSettings.DefaultBodyBuffSize)
		httpParser := parser.NewHTTPParser(&request, pipe, serverSettings)

		handler := server.NewHTTPHandler(server.HTTPHandlerArgs{
			Router:  router,
			Request: &request,
			Parser:  httpParser,
			RespWriter: func(b []byte) error {
				_, err := conn.Write(b)
				return err
			},
		})

		go handler.Poll()
		server.DefaultConnHandler(wg, conn, handler.OnData)
	}, shutdown)
}

func getSettings(maybeSettings []settings.Settings) (settings.Settings, error) {
	switch len(maybeSettings) {
	case 0:
		return settings.Default(), nil
	case 1:
		return settings.Prepare(maybeSettings[0]), nil
	default:
		return settings.Settings{}, errors.ErrTooMuchSettings
	}
}
