package indigo

import (
	"fmt"
	"indigo/httpparser"
	"indigo/httpserver"
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

	return httpserver.StartTCPServer(sock, func(conn net.Conn) {
		request, writeBody := types.NewRequest(make([]byte, 10), make(map[string][]byte), nil)
		parser := httpparser.NewHTTPParser(&request, writeBody, httpparser.Settings{})

		handler := httpserver.NewHTTPHandler(httpserver.HTTPHandlerArgs{
			Router:           router,
			Request:          &request,
			WriteRequestBody: writeBody,
			Parser:           parser,
			RespWriter: func(b []byte) error {
				_, err = conn.Write(b)
				return err
			},
		})

		go handler.Poll()
		httpserver.DefaultConnHandler(conn, handler.OnData)
	})
}
