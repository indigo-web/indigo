package construct

import (
	"net"

	"github.com/indigo-web/indigo/config"
	"github.com/indigo-web/indigo/http"
	"github.com/indigo-web/indigo/internal/buffer"
	"github.com/indigo-web/indigo/kv"
	"github.com/indigo-web/indigo/transport"
)

func Request(cfg *config.Config, client transport.Client) *http.Request {
	headers := kv.NewPrealloc(int(cfg.Headers.Number.Default))
	params := kv.NewPrealloc(cfg.URI.ParamsPrealloc)
	vars := kv.New()
	request := http.NewRequest(cfg, http.NewResponse(), client, headers, params, vars)

	return request
}

func Client(cfg config.NET, conn net.Conn) transport.Client {
	readBuff := make([]byte, cfg.ReadBufferSize)

	return transport.NewClient(conn, cfg.ReadTimeout, readBuff)
}

func Buffers(s *config.Config) (headersBuff, statusBuff *buffer.Buffer) {
	return buffer.New(s.Headers.Space.Default, s.Headers.Space.Maximal),
		buffer.New(s.URI.RequestLineSize.Default, s.URI.RequestLineSize.Maximal)
}
