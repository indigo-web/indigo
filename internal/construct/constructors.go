package construct

import (
	"github.com/indigo-web/chunkedbody"
	"github.com/indigo-web/indigo/config"
	"github.com/indigo-web/indigo/http"
	"github.com/indigo-web/indigo/http/headers"
	"github.com/indigo-web/indigo/http/query"
	"github.com/indigo-web/indigo/internal/keyvalue"
	"github.com/indigo-web/indigo/internal/tcp"
	"github.com/indigo-web/utils/buffer"
	"net"
)

func Request(cfg *config.Config, client tcp.Client, body http.Retriever) *http.Request {
	hdrs := headers.NewPrealloc(cfg.Headers.Number.Default)
	q := query.New(keyvalue.NewPreAlloc(cfg.URL.Query.PreAlloc))
	resp := http.NewResponse()
	params := keyvalue.New()

	return http.NewRequest(cfg, hdrs, q, resp, client, http.NewBody(body, cfg), params)
}

func Chunked(cfg config.Body) *chunkedbody.Parser {
	return chunkedbody.NewParser(chunkedbody.Settings{
		MaxChunkSize: cfg.MaxChunkSize,
	})
}

func Client(cfg config.TCP, conn net.Conn) tcp.Client {
	readBuff := make([]byte, cfg.ReadBufferSize)

	return tcp.NewClient(conn, cfg.ReadTimeout, readBuff)
}

func Buffers(s *config.Config) (keyBuff *buffer.Buffer, valBuff *buffer.Buffer, startLineBuff *buffer.Buffer) {
	return buffer.New(s.Headers.KeySpace.Default, s.Headers.KeySpace.Maximal),
		buffer.New(s.Headers.ValueSpace.Default, s.Headers.ValueSpace.Maximal),
		buffer.New(s.URL.BufferSize.Default, s.URL.BufferSize.Maximal)
}
