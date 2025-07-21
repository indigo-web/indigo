package construct

import (
	"github.com/indigo-web/indigo/config"
	"github.com/indigo-web/indigo/http"
	"github.com/indigo-web/indigo/internal/buffer"
	"github.com/indigo-web/indigo/kv"
	"github.com/indigo-web/indigo/transport"
	"net"
)

func Request(cfg *config.Config, client transport.Client) *http.Request {
	headers := kv.NewPrealloc(cfg.Headers.Number.Default)
	params := kv.NewPrealloc(cfg.URI.ParamsPrealloc)
	vars := kv.New()
	request := http.NewRequest(cfg, http.NewResponse(), client, headers, params, vars)

	return request
}

func Client(cfg config.NET, conn net.Conn) transport.Client {
	readBuff := make([]byte, cfg.ReadBufferSize)

	return transport.NewClient(conn, cfg.ReadTimeout, readBuff)
}

func Buffers(s *config.Config) (keysBuff, valsBuff, statusBuff buffer.Buffer) {
	return buffer.New(s.Headers.KeySpace.Default, s.Headers.KeySpace.Maximal),
		buffer.New(s.Headers.ValueSpace.Default, s.Headers.ValueSpace.Maximal),
		buffer.New(s.URI.RequestLineSize.Default, s.URI.RequestLineSize.Maximal)
}
