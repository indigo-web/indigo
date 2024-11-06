package serve

import (
	"github.com/indigo-web/indigo/config"
	"github.com/indigo-web/indigo/http/crypt"
	"github.com/indigo-web/indigo/internal/construct"
	"github.com/indigo-web/indigo/internal/protocol/http1"
	"github.com/indigo-web/indigo/router"
	"net"
)

// HTTP1 setups and serves an HTTP/1.1 server until it stops. Note, that the connection isn't
// automatically closed on server stop
func HTTP1(cfg *config.Config, conn net.Conn, enc crypt.Encryption, r router.Router) {
	client := construct.Client(cfg.NET, conn)
	body := http1.NewBody(client, construct.Chunked(cfg.Body), cfg.Body)
	request := construct.Request(cfg, client, body)
	request.Env.Encryption = enc
	suit := http1.Initialize(cfg, r, client, request, body)
	suit.Serve()
}
