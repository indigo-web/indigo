package indigo

import (
	"github.com/indigo-web/chunkedbody"
	"github.com/indigo-web/indigo/http"
	"github.com/indigo-web/indigo/http/coding"
	"github.com/indigo-web/indigo/http/headers"
	"github.com/indigo-web/indigo/http/query"
	"github.com/indigo-web/indigo/internal/server/tcp"
	"github.com/indigo-web/indigo/internal/transport"
	"github.com/indigo-web/indigo/internal/transport/http1"
	"github.com/indigo-web/indigo/settings"
	"github.com/indigo-web/utils/buffer"
	"github.com/indigo-web/utils/pool"
	"net"
)

func newClient(tcpSettings settings.TCP, conn net.Conn) tcp.Client {
	readBuff := make([]byte, tcpSettings.ReadBufferSize)

	return tcp.NewClient(conn, tcpSettings.ReadTimeout, readBuff)
}

func newKeyValueBuffs(s settings.Headers) (*buffer.Buffer[byte], *buffer.Buffer[byte]) {
	keyBuff := buffer.NewBuffer[byte](
		s.MaxKeyLength*s.Number.Default,
		s.MaxKeyLength*s.Number.Maximal,
	)
	valBuff := buffer.NewBuffer[byte](
		s.ValueSpace.Default,
		s.ValueSpace.Maximal,
	)

	return keyBuff, valBuff
}

func newBody(client tcp.Client, body settings.Body, codings []coding.Constructor) http.Body {
	manager := coding.NewManager(body.DecodedBufferSize)

	for _, constructor := range codings {
		manager.AddCoding(constructor)
	}

	chunkedBodySettings := chunkedbody.DefaultSettings()
	chunkedBodySettings.MaxChunkSize = body.MaxChunkSize

	return http1.NewBody(client, chunkedbody.NewParser(chunkedBodySettings), manager)
}

func newRequest(s settings.Settings, conn net.Conn, body http.Body) *http.Request {
	q := query.NewQuery(headers.NewHeaders())
	hdrs := headers.NewPreallocHeaders(s.Headers.Number.Default)
	response := http.NewResponse()
	params := make(http.Params)

	return http.NewRequest(hdrs, q, response, conn, body, params, s.URL.Params.DisableMapClear)
}

func newTransport(s settings.Settings, req *http.Request, a *Application) transport.Transport {
	keyBuff, valBuff := newKeyValueBuffs(s.Headers)
	objPool := pool.NewObjectPool[[]string](s.Headers.MaxValuesObjectPoolSize)
	startLineBuff := buffer.NewBuffer[byte](
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
		nil, // TODO: pass response buff size from settings,
		a.defaultHeaders,
	)
}
