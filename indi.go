package indigo

import (
	"fmt"
	"indigo/http/parser"
	"indigo/http/server"
	"indigo/router"
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

func (a Application) Serve(router router.Router) error {
	address := fmt.Sprintf("%s:%d", a.addr, a.port)
	sock, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}
	defer sock.Close()

	return server.StartTCPServer(sock, func(conn net.Conn) {
		request, pipe := types.NewRequest(make([]byte, 10), make(map[string][]byte), nil)
		parser := parser.NewHTTPParser(&request, pipe, parser.Settings{})

		handler := server.NewHTTPHandler(server.HTTPHandlerArgs{
			Router:  router,
			Request: &request,
			Parser:  parser,
			RespWriter: func(b []byte) error {
				_, err = conn.Write(b)
				return err
			},
		})

		go handler.Poll()
		server.DefaultConnHandler(conn, handler.OnData)
	})
}
