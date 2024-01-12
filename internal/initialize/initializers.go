package initialize

import (
	"github.com/indigo-web/chunkedbody"
	"github.com/indigo-web/indigo/http"
	"github.com/indigo-web/indigo/http/headers"
	"github.com/indigo-web/indigo/http/query"
	"github.com/indigo-web/indigo/internal/keyvalue"
	"github.com/indigo-web/indigo/internal/server/tcp"
	"github.com/indigo-web/indigo/internal/transport"
	"github.com/indigo-web/indigo/internal/transport/http1"
	"github.com/indigo-web/indigo/settings"
	"github.com/indigo-web/utils/buffer"
	"github.com/indigo-web/utils/pool"
	"net"
)

func NewClient(tcpSettings settings.TCP, conn net.Conn) tcp.Client {
	readBuff := make([]byte, tcpSettings.ReadBufferSize)

	return tcp.NewClient(conn, tcpSettings.ReadTimeout, readBuff)
}

func NewKeyValueBuffs(s settings.Headers) (*buffer.Buffer, *buffer.Buffer) {
	keyBuff := buffer.New(
		s.MaxKeyLength*s.Number.Default,
		s.MaxKeyLength*s.Number.Maximal,
	)
	valBuff := buffer.New(
		s.ValueSpace.Default,
		s.ValueSpace.Maximal,
	)

	return keyBuff, valBuff
}

func NewBody(client tcp.Client, body settings.Body) http.Body {
	chunkedBodySettings := chunkedbody.DefaultSettings()
	chunkedBodySettings.MaxChunkSize = body.MaxChunkSize

	return http1.NewBody(client, chunkedbody.NewParser(chunkedBodySettings), body)
}

func NewRequest(s settings.Settings, conn net.Conn, body http.Body) *http.Request {
	q := query.NewQuery(headers.NewPrealloc(s.URL.Query.PreAlloc))
	hdrs := headers.NewPrealloc(s.Headers.Number.Default)
	response := http.NewResponse()
	params := keyvalue.New()

	return http.NewRequest(hdrs, q, response, conn, body, params)
}

func NewTransport(s settings.Settings, req *http.Request) transport.Transport {
	keyBuff, valBuff := NewKeyValueBuffs(s.Headers)
	objPool := pool.NewObjectPool[[]string](s.Headers.MaxValuesObjectPoolSize)
	startLineBuff := buffer.New(
		s.URL.BufferSize.Default,
		s.URL.BufferSize.Maximal,
	)
	respBuff := make([]byte, 0, s.HTTP.ResponseBuffSize)

	return http1.New(
		req,
		*keyBuff, *valBuff, *startLineBuff,
		*objPool,
		s.Headers,
		respBuff,
		s.HTTP.FileBuffSize,
		s.Headers.Default,
	)
}
