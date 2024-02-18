package core

import (
	"github.com/indigo-web/indigo/config"
	"github.com/indigo-web/indigo/http/encryption"
	"github.com/indigo-web/indigo/internal/initialize"
	"github.com/indigo-web/indigo/internal/server/http"
	"github.com/indigo-web/indigo/router"
	"net"
)

func ServeHTTP1(s config.Config, conn net.Conn, enc encryption.Token, r router.Router) {
	client := initialize.NewClient(s.TCP, conn)
	body := initialize.NewBody(client, s.Body)
	request := initialize.NewRequest(s, conn, body)
	request.Env.Encryption = enc
	trans := initialize.NewTransport(s, request)
	httpServer := http.NewServer(r, s.HTTP.OnDisconnect)
	httpServer.Run(client, request, trans)
}
