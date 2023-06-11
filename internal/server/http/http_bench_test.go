package http

import (
	"github.com/indigo-web/indigo/http/decode"
	"github.com/indigo-web/indigo/http/method"
	"github.com/indigo-web/indigo/http/status"
	"github.com/indigo-web/indigo/internal/server/tcp/dummy"
	"github.com/indigo-web/indigo/router"
	"github.com/indigo-web/indigo/router/inbuilt"
	"github.com/indigo-web/indigo/router/simple"
	"testing"

	"github.com/indigo-web/indigo/http"
	"github.com/indigo-web/indigo/internal/pool"

	"github.com/indigo-web/indigo/http/headers"
	"github.com/indigo-web/indigo/http/query"
	"github.com/indigo-web/indigo/internal/arena"
	"github.com/indigo-web/indigo/internal/parser/http1"
	render2 "github.com/indigo-web/indigo/internal/render"
	"github.com/indigo-web/indigo/settings"
)

var (
	simpleGETRequest      = []byte("GET / HTTP/1.1\r\n\r\n")
	fiveHeadersGETRequest = []byte(
		"GET / HTTP/1.1\r\n" +
			"Hello: world\r\n" +
			"One: ok\r\n" +
			"Content-Type: nothing but true;q=0.9\r\n" +
			"Four: lorem ipsum\r\n" +
			"Mistake: is made here\r\n" +
			"\r\n",
	)
	tenHeadersGETRequest = []byte(
		"GET / HTTP/1.1\r\n" +
			"Hello: world\r\n" +
			"One: ok\r\n" +
			"Content-Type: nothing but true;q=0.9\r\n" +
			"Four: lorem ipsum\r\n" +
			"Mistake: is made here\r\n" +
			"Lorem: upsum\r\n" +
			"tired: of all this shit\r\n" +
			"Eight: finally only two left\r\n" +
			"my-brain: is not so creative\r\n" +
			"to-create: ten random headers from scratch\r\n" +
			"\r\n",
	)
	simpleGETWithHeader = []byte("GET /with-header HTTP/1.1\r\n\r\n")

	simplePOST = []byte("POST / HTTP/1.1\r\nContent-Length: 13\r\n\r\nHello, world!")
)

// Using default headers pasted from indi.go. Not using original ones as
// this leads to cycle import (why cannot compiler handle such situations?)
var defaultHeaders = map[string][]string{
	"Content-Type": {"text/html"},
	// nil here means that value will be set later, when server will be initializing
	"Accept-Encodings": nil,
}

func getInbuiltRouter() router.Router {
	r := inbuilt.NewRouter()
	root := r.Resource("/")
	root.Get(http.RespondTo)
	root.Post(func(request *http.Request) http.Response {
		_ = request.OnBody(func([]byte) error {
			return nil
		}, func() error {
			return nil
		})

		return http.RespondTo(request)
	})
	r.Get("/with-header", func(request *http.Request) http.Response {
		return http.RespondTo(request).WithHeader("Hello", "World")
	})
	r.OnStart()

	return r
}

func getSimpleRouter() router.Router {
	r := simple.NewRouter(func(request *http.Request) http.Response {
		switch request.Path.String {
		case "/":
			switch request.Method {
			case method.GET:
				return http.RespondTo(request)
			case method.POST:
				_ = request.OnBody(func([]byte) error {
					return nil
				}, func() error {
					return nil
				})

				return http.RespondTo(request)
			default:
				return http.RespondTo(request).WithError(status.ErrMethodNotAllowed)
			}
		case "/with-header":
			return http.RespondTo(request).WithHeader("Hello", "World")
		default:
			return http.RespondTo(request).
				WithError(status.ErrNotFound)
		}
	}, func(request *http.Request, err error) http.Response {
		return http.RespondTo(request).WithError(err)
	})

	return r
}

func Benchmark_Get(b *testing.B) {
	// for benchmarking, using more realistic conditions. In case we want a pure performance - use
	// getSimpleRouter() here. It is visibly faster
	r := getInbuiltRouter()
	s := settings.Default()
	q := query.NewQuery(func() query.Map {
		return make(query.Map)
	})
	bodyReader := http1.NewBodyReader(dummy.NewNopClient(), s.Body, decode.NewDecoder())
	hdrs := headers.NewHeaders(make(map[string][]string, 10))
	request := http.NewRequest(
		hdrs, q, http.NewResponse(), dummy.NewNopConn(), bodyReader, nil, false,
	)
	keyArena := arena.NewArena(
		s.Headers.MaxKeyLength*s.Headers.Number.Default,
		s.Headers.MaxKeyLength*s.Headers.Number.Maximal,
	)
	valArena := arena.NewArena(
		s.Headers.ValueSpace.Default, s.Headers.ValueSpace.Maximal,
	)
	objPool := pool.NewObjectPool[[]string](20)
	startLineBuff := make([]byte, s.URL.MaxLength)
	parser := http1.NewHTTPRequestsParser(
		request, keyArena, valArena, objPool, startLineBuff, s.Headers,
	)
	render := render2.NewEngine(make([]byte, 0, 1024), nil, defaultHeaders)
	server := NewHTTPServer(r).(*httpServer)

	b.Run("SimpleGET", func(b *testing.B) {
		simpleGETClient := dummy.NewCircularClient(simpleGETRequest)
		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			server.RunOnce(simpleGETClient, request, bodyReader, render, parser)
		}
	})

	b.Run("FiveHeadersGET", func(b *testing.B) {
		fiveHeadersGETClient := dummy.NewCircularClient(fiveHeadersGETRequest)
		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			server.RunOnce(fiveHeadersGETClient, request, bodyReader, render, parser)
		}
	})

	b.Run("TenHeadersGET", func(b *testing.B) {
		tenHeadersGETClient := dummy.NewCircularClient(tenHeadersGETRequest)
		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			server.RunOnce(tenHeadersGETClient, request, bodyReader, render, parser)
		}
	})

	b.Run("WithRespHeader", func(b *testing.B) {
		withRespHeadersGETClient := dummy.NewCircularClient(simpleGETWithHeader)
		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			server.RunOnce(withRespHeadersGETClient, request, bodyReader, render, parser)
		}
	})
}

func Benchmark_Post(b *testing.B) {
	r := getInbuiltRouter()
	s := settings.Default()
	q := query.NewQuery(func() query.Map {
		return make(query.Map)
	})
	hdrs := headers.NewHeaders(make(map[string][]string, 10))
	withBodyClient := dummy.NewCircularClient(simplePOST)
	reader := http1.NewBodyReader(withBodyClient, settings.Default().Body, decode.NewDecoder())
	request := http.NewRequest(
		hdrs, q, http.NewResponse(), dummy.NewNopConn(), reader, nil, false,
	)
	keyArena := arena.NewArena(
		s.Headers.MaxKeyLength*s.Headers.Number.Default,
		s.Headers.MaxKeyLength*s.Headers.Number.Maximal,
	)
	valArena := arena.NewArena(
		s.Headers.ValueSpace.Default, s.Headers.ValueSpace.Maximal,
	)
	objPool := pool.NewObjectPool[[]string](20)
	startLineBuff := make([]byte, s.URL.MaxLength)
	parser := http1.NewHTTPRequestsParser(
		request, keyArena, valArena, objPool, startLineBuff, s.Headers,
	)
	render := render2.NewEngine(make([]byte, 0, 1024), nil, defaultHeaders)
	server := NewHTTPServer(r).(*httpServer)

	b.Run("SimplePOST", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			server.RunOnce(withBodyClient, request, reader, render, parser)
		}
	})
}
