package indigo

import (
	"errors"
	"fmt"
	"indigo/http/parser"
	"indigo/http/server"
	"indigo/router"
	"indigo/settings"
	"indigo/types"
	"net"
)

// Application TODO: add domain to allow host multiple
//                   domains in a single web-server instance
type Application struct {
	addr string
	port uint16
}

func NewApp(addr string, port uint16) *Application {
	return &Application{
		addr: addr,
		port: port,
	}
}

func (a Application) Serve(router router.Router, maybeSettings ...settings.Settings) error {
	var serverSettings settings.Settings

	switch len(maybeSettings) {
	case 0:
		serverSettings = settings.Default()
	case 1:
		serverSettings = settings.Prepare(maybeSettings[0])
	default:
		return errors.New("too much settings (one struct is expected)")
	}

	address := fmt.Sprintf("%s:%d", a.addr, a.port)
	sock, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}
	defer sock.Close()

	return server.StartTCPServer(sock, func(conn net.Conn) {
		request, pipe := types.NewRequest(make([]byte, 10), make(map[string][]byte), nil)
		httpParser := parser.NewHTTPParser(&request, pipe, serverSettings)

		handler := server.NewHTTPHandler(server.HTTPHandlerArgs{
			Router:  router,
			Request: &request,
			Parser:  httpParser,
			RespWriter: func(b []byte) error {
				_, err = conn.Write(b)
				return err
			},
		})

		go handler.Poll()
		server.DefaultConnHandler(conn, handler.OnData)
	})
}
