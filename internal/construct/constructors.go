package construct

import (
	"github.com/indigo-web/chunkedbody"
	"github.com/indigo-web/indigo/config"
	"github.com/indigo-web/indigo/http"
	"github.com/indigo-web/indigo/http/headers"
	"github.com/indigo-web/indigo/http/query"
	"github.com/indigo-web/indigo/internal/keyvalue"
	"github.com/indigo-web/indigo/transport"
	"github.com/indigo-web/utils/buffer"
	"net"
)

func Request(cfg *config.Config, client transport.Client, body http.Retriever) *http.Request {
	hdrs := headers.NewPrealloc(cfg.Headers.Number.Default)
	q := query.New(keyvalue.NewPreAlloc(cfg.URL.Query.ParamsPrealloc), cfg)
	resp := http.NewResponse()
	params := keyvalue.New()
	request := http.NewRequest(cfg, hdrs, q, resp, client, params)
	request.Body = http.NewBody(request, body, cfg)

	return request
}

func Chunked(cfg config.Body) *chunkedbody.Parser {
	return chunkedbody.NewParser(chunkedbody.Settings{
		MaxChunkSize: int64(cfg.MaxChunkSize),
	})
}

func Client(cfg config.NET, conn net.Conn) transport.Client {
	readBuff := make([]byte, cfg.ReadBufferSize)

	return transport.NewClient(conn, cfg.ReadTimeout, readBuff)
}

func Buffers(s *config.Config) (keyBuff *buffer.Buffer, valBuff *buffer.Buffer, startLineBuff *buffer.Buffer) {
	return buffer.New(s.Headers.KeySpace.Default, s.Headers.KeySpace.Maximal),
		buffer.New(s.Headers.ValueSpace.Default, s.Headers.ValueSpace.Maximal),
		buffer.New(s.URL.BufferSize.Default, s.URL.BufferSize.Maximal)
}
