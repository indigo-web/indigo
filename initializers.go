package indigo

import (
	"github.com/indigo-web/indigo/http"
	"github.com/indigo-web/indigo/http/decode"
	"github.com/indigo-web/indigo/http/headers"
	"github.com/indigo-web/indigo/http/query"
	httpparser "github.com/indigo-web/indigo/internal/parser"
	"github.com/indigo-web/indigo/internal/parser/http1"
	"github.com/indigo-web/indigo/internal/render"
	"github.com/indigo-web/indigo/internal/server/tcp"
	"github.com/indigo-web/indigo/settings"
	"github.com/indigo-web/utils/arena"
	"github.com/indigo-web/utils/pool"
	"net"
)

func newClient(tcpSettings settings.TCP, conn net.Conn) tcp.Client {
	readBuff := make([]byte, tcpSettings.ReadBufferSize)

	return tcp.NewClient(conn, tcpSettings.ReadTimeout, readBuff)
}

func newKeyValueArenas(s settings.Headers) (arena.Arena, arena.Arena) {
	keyArena := arena.NewArena(
		s.MaxKeyLength*s.Number.Default,
		s.MaxKeyLength*s.Number.Maximal,
	)
	valArena := arena.NewArena(
		s.ValueSpace.Default,
		s.ValueSpace.Maximal,
	)

	return keyArena, valArena
}

func newRequest(
	s settings.Settings, conn net.Conn, r http.BodyReader, decoder *decode.Decoder,
) *http.Request {
	q := query.NewQuery(func() query.Map {
		return make(query.Map, s.URL.Query.DefaultMapSize)
	})
	hdrs := headers.NewHeaders(make(map[string][]string, s.Headers.Number.Default))
	response := http.NewResponse()
	params := make(http.Params)
	body := http.NewBody(r, decoder)

	return http.NewRequest(hdrs, q, response, conn, body, params, s.URL.Params.DisableMapClear)
}

func newRenderer(httpSettings settings.HTTP, a *Application) render.Engine {
	respBuff := make([]byte, 0, httpSettings.ResponseBuffSize)

	return render.NewEngine(respBuff, nil, a.defaultHeaders)
}

func newHTTPParser(s settings.Settings, req *http.Request) httpparser.HTTPRequestsParser {
	keyAlloc, valAlloc := newKeyValueArenas(s.Headers)
	objPool := pool.NewObjectPool[[]string](s.Headers.MaxValuesObjectPoolSize)

	startLineBuff := make([]byte, s.URL.MaxLength)

	return http1.NewHTTPRequestsParser(
		req, keyAlloc, valAlloc, objPool, startLineBuff, s.Headers,
	)
}
