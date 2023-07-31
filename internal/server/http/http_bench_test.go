package http

import (
	"github.com/indigo-web/indigo/http/decode"
	"github.com/indigo-web/indigo/http/method"
	"github.com/indigo-web/indigo/http/status"
	"github.com/indigo-web/indigo/internal/requestgen"
	"github.com/indigo-web/indigo/internal/server/tcp/dummy"
	"github.com/indigo-web/indigo/router"
	"github.com/indigo-web/indigo/router/inbuilt"
	"github.com/indigo-web/indigo/router/simple"
	"strings"
	"testing"

	"github.com/indigo-web/indigo/http"
	"github.com/indigo-web/utils/pool"

	"github.com/indigo-web/indigo/http/headers"
	"github.com/indigo-web/indigo/http/query"
	"github.com/indigo-web/indigo/internal/parser/http1"
	render2 "github.com/indigo-web/indigo/internal/render"
	"github.com/indigo-web/indigo/settings"
	"github.com/indigo-web/utils/arena"
)

var (
	tenHeadersGETRequest = []byte(
		"GET / HTTP/1.1\r\n" +
			"Hello: world\r\n" +
			"One: ok\r\n" +
			"Content-Type: nothing but true;q=0.9\r\n" +
			"Four: lorem ipsum\r\n" +
			"Mistake: is made here\r\n" +
			"Lorem: ipsum\r\n" +
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
		_ = request.Body().Callback(func([]byte) error {
			return nil
		})

		return http.RespondTo(request)
	})
	r.Get("/with-header", func(request *http.Request) http.Response {
		return http.RespondTo(request).WithHeader("Hello", "World")
	})

	if err := r.OnStart(); err != nil {
		panic(err)
	}

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
				_ = request.Body().Callback(func([]byte) error {
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
	q := query.NewQuery(headers.NewHeaders(nil))
	bodyReader := http1.NewBodyReader(dummy.NewNopClient(), s.Body)
	body := http.NewBody(bodyReader, decode.NewDecoder())
	hdrs := headers.NewHeaders(make(map[string][]string, 10))
	request := http.NewRequest(
		hdrs, q, http.NewResponse(), dummy.NewNopConn(), http.NewBody(bodyReader, decode.NewDecoder()),
		nil, false,
	)
	keyArena := arena.NewArena[byte](
		s.Headers.MaxKeyLength*s.Headers.Number.Default,
		s.Headers.MaxKeyLength*s.Headers.Number.Maximal,
	)
	valArena := arena.NewArena[byte](
		s.Headers.ValueSpace.Default, s.Headers.ValueSpace.Maximal,
	)
	objPool := pool.NewObjectPool[[]string](20)
	startLineArena := arena.NewArena[byte](
		s.URL.BufferSize.Default,
		s.URL.BufferSize.Maximal,
	)
	parser := http1.NewHTTPRequestsParser(
		request, *keyArena, *valArena, *startLineArena, *objPool, s.Headers,
	)
	render := render2.NewEngine(make([]byte, 0, 1024), nil, defaultHeaders)
	server := NewHTTPServer(r).(*httpServer)

	b.Run("simple get", func(b *testing.B) {
		tenHeadersGETClient := dummy.NewCircularClient(tenHeadersGETRequest)
		b.SetBytes(int64(len(tenHeadersGETRequest)))
		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			server.RunOnce(tenHeadersGETClient, request, body, render, parser)
		}
	})

	b.Run("with resp header", func(b *testing.B) {
		withRespHeadersGETClient := dummy.NewCircularClient(simpleGETWithHeader)
		b.SetBytes(int64(len(simpleGETWithHeader)))
		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			server.RunOnce(withRespHeadersGETClient, request, body, render, parser)
		}
	})

	b.Run("5 headers", func(b *testing.B) {
		data := requestgen.Generate(strings.Repeat("a", 500), 5)
		client := dummy.NewCircularClient(data)
		b.SetBytes(int64(len(data)))
		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			server.RunOnce(client, request, body, render, parser)
		}
	})

	b.Run("10 headers", func(b *testing.B) {
		data := requestgen.Generate(strings.Repeat("a", 500), 10)
		client := dummy.NewCircularClient(data)
		b.SetBytes(int64(len(data)))
		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			server.RunOnce(client, request, body, render, parser)
		}
	})

	b.Run("50 headers", func(b *testing.B) {
		data := requestgen.Generate(strings.Repeat("a", 500), 50)
		client := dummy.NewCircularClient(data)
		b.SetBytes(int64(len(data)))
		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			server.RunOnce(client, request, body, render, parser)
		}
	})

	b.Run("heavily escaped", func(b *testing.B) {
		data := requestgen.Generate(strings.Repeat("%20", 500), 20)
		client := dummy.NewCircularClient(data)
		b.SetBytes(int64(len(data)))
		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			server.RunOnce(client, request, body, render, parser)
		}
	})
}

func Benchmark_Post(b *testing.B) {
	r := getInbuiltRouter()
	s := settings.Default()
	q := query.NewQuery(headers.NewHeaders(nil))
	hdrs := headers.NewHeaders(make(map[string][]string, 10))
	withBodyClient := dummy.NewCircularClient(simplePOST)
	reader := http1.NewBodyReader(withBodyClient, settings.Default().Body)
	body := http.NewBody(reader, decode.NewDecoder())
	request := http.NewRequest(
		hdrs, q, http.NewResponse(), dummy.NewNopConn(), http.NewBody(reader, decode.NewDecoder()),
		nil, false,
	)
	keyArena := arena.NewArena[byte](
		s.Headers.MaxKeyLength*s.Headers.Number.Default,
		s.Headers.MaxKeyLength*s.Headers.Number.Maximal,
	)
	valArena := arena.NewArena[byte](
		s.Headers.ValueSpace.Default, s.Headers.ValueSpace.Maximal,
	)
	objPool := pool.NewObjectPool[[]string](20)
	startLineArena := arena.NewArena[byte](
		s.URL.BufferSize.Default,
		s.URL.BufferSize.Maximal,
	)
	parser := http1.NewHTTPRequestsParser(
		request, *keyArena, *valArena, *startLineArena, *objPool, s.Headers,
	)
	render := render2.NewEngine(make([]byte, 0, 1024), nil, defaultHeaders)
	server := NewHTTPServer(r).(*httpServer)

	b.Run("SimplePOST", func(b *testing.B) {
		b.SetBytes(int64(len(simplePOST)))
		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			server.RunOnce(withBodyClient, request, body, render, parser)
		}
	})
}
