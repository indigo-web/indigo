package indigo

import (
	"indigo/http/headers"
	"indigo/http/parser/http1"
	"indigo/http/server"
	"indigo/router"
	"indigo/types"
	"net"
)

type Application struct {
	addr string

	sock *net.Listener

	shutdown chan bool
}

func NewApp(addr string) Application {
	return Application{
		addr:     addr,
		shutdown: make(chan bool),
	}
}

func (a Application) Serve(router router.Router) error {
	sock, err := net.Listen("tcp", a.addr)
	if err != nil {
		return err
	}

	return server.StartTCPServer(sock, func(conn net.Conn) {
		// TODO: add settings and add this value right there
		headersManager := headers.NewManager(5)
		request, gateway := types.NewRequest(headersManager)
		// TODO: add startLine and headerBuff initial size to settings
		startLineBuff := make([]byte, 0, 1024)
		headerBuff := make([]byte, 0, 100)
		httpParser := http1.NewHTTPRequestsParser(
			request, startLineBuff, headerBuff, gateway,
		)
		httpServer := server.NewHTTPServer(request, func(b []byte) error {
			_, err := conn.Write(b)
			return err
		}, router, httpParser)
		go httpServer.Run()

		readBuff := make([]byte, 1024)
		server.DefaultConnHandler(conn, readBuff, httpServer.OnData)
	})
}
