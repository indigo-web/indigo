package http

import (
	"context"
	"github.com/indigo-web/indigo/http/coding"
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
	"github.com/indigo-web/indigo/internal/render"
	"github.com/indigo-web/indigo/settings"
	"github.com/indigo-web/utils/buffer"
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

var longPath = strings.Repeat("a", 500)

func getInbuiltRouter() router.Router {
	r := inbuilt.New().
		Get("/with-header", func(request *http.Request) *http.Response {
			return request.Respond().WithHeader("Hello", "World")
		}).
		Get("/"+longPath, http.Respond)

	r.Resource("/").
		Get(http.Respond).
		Post(func(request *http.Request) *http.Response {
			_ = request.Body.Callback(func([]byte) error {
				return nil
			})

			return request.Respond()
		})

	if err := r.OnStart(); err != nil {
		panic(err)
	}

	return r
}

func getSimpleRouter() router.Router {
	longpath := "/" + longPath

	r := simple.NewRouter(func(request *http.Request) *http.Response {
		switch request.Path {
		case "/":
			switch request.Method {
			case method.GET:
				return request.Respond()
			case method.POST:
				_ = request.Body.Callback(func([]byte) error {
					return nil
				})

				return request.Respond()
			default:
				return request.Respond().WithError(status.ErrMethodNotAllowed)
			}
		case "/with-header":
			return request.Respond().WithHeader("Hello", "World")
		case longpath:
			return request.Respond()
		default:
			return request.Respond().
				WithError(status.ErrNotFound)
		}
	}, func(request *http.Request, err error) *http.Response {
		return request.Respond().WithError(err)
	})

	return r
}

func Benchmark_Get(b *testing.B) {
	// for benchmarking, using more realistic conditions. In case we want a pure performance - use
	// getSimpleRouter() here. It is visibly faster
	r := getInbuiltRouter()
	s := settings.Default()
	q := query.NewQuery(headers.NewHeaders())
	body := http1.NewBody(
		dummy.NewNopClient(), nil, coding.NewManager(0),
	)
	hdrs := headers.FromMap(make(map[string][]string, 10))
	request := http.NewRequest(
		context.Background(), hdrs, q, http.NewResponse(), dummy.NewNopConn(), body,
		nil, false,
	)
	keyArena := buffer.NewBuffer[byte](
		s.Headers.MaxKeyLength*s.Headers.Number.Default,
		s.Headers.MaxKeyLength*s.Headers.Number.Maximal,
	)
	valArena := buffer.NewBuffer[byte](
		s.Headers.ValueSpace.Default, s.Headers.ValueSpace.Maximal,
	)
	objPool := pool.NewObjectPool[[]string](20)
	startLineArena := buffer.NewBuffer[byte](
		s.URL.BufferSize.Default,
		s.URL.BufferSize.Maximal,
	)
	parser := http1.NewHTTPRequestsParser(
		request, *keyArena, *valArena, *startLineArena, *objPool, s.Headers,
	)
	renderer := render.NewEngine(make([]byte, 0, 1024), nil, defaultHeaders)
	server := NewHTTPServer(r).(*httpServer)

	b.Run("simple get", func(b *testing.B) {
		tenHeadersGETClient := dummy.NewCircularClient(tenHeadersGETRequest)
		b.SetBytes(int64(len(tenHeadersGETRequest)))
		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			server.RunOnce(tenHeadersGETClient, request, renderer, parser)
		}
	})

	b.Run("with resp header", func(b *testing.B) {
		withRespHeadersGETClient := dummy.NewCircularClient(simpleGETWithHeader)
		b.SetBytes(int64(len(simpleGETWithHeader)))
		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			server.RunOnce(withRespHeadersGETClient, request, renderer, parser)
		}
	})

	b.Run("5 headers", func(b *testing.B) {
		data := requestgen.Generate(longPath, 5)
		client := dummy.NewCircularClient(data)
		b.SetBytes(int64(len(data)))
		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			server.RunOnce(client, request, renderer, parser)
		}
	})

	b.Run("10 headers", func(b *testing.B) {
		data := requestgen.Generate(longPath, 10)
		client := dummy.NewCircularClient(data)
		b.SetBytes(int64(len(data)))
		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			server.RunOnce(client, request, renderer, parser)
		}
	})

	b.Run("50 headers", func(b *testing.B) {
		data := requestgen.Generate(longPath, 50)
		client := dummy.NewCircularClient(data)
		b.SetBytes(int64(len(data)))
		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			server.RunOnce(client, request, renderer, parser)
		}
	})

	b.Run("heavily escaped", func(b *testing.B) {
		data := requestgen.Generate(strings.Repeat("%20", 500), 20)
		client := dummy.NewCircularClient(data)
		b.SetBytes(int64(len(data)))
		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			server.RunOnce(client, request, renderer, parser)
		}
	})
}

func Benchmark_Post(b *testing.B) {
	r := getInbuiltRouter()
	s := settings.Default()
	q := query.NewQuery(headers.NewHeaders())
	hdrs := headers.FromMap(make(map[string][]string, 10))
	withBodyClient := dummy.NewCircularClient(simplePOST)
	body := http1.NewBody(
		withBodyClient, nil, coding.NewManager(0),
	)
	request := http.NewRequest(
		context.Background(), hdrs, q, http.NewResponse(), dummy.NewNopConn(), body,
		nil, false,
	)
	keyArena := buffer.NewBuffer[byte](
		s.Headers.MaxKeyLength*s.Headers.Number.Default,
		s.Headers.MaxKeyLength*s.Headers.Number.Maximal,
	)
	valArena := buffer.NewBuffer[byte](
		s.Headers.ValueSpace.Default, s.Headers.ValueSpace.Maximal,
	)
	objPool := pool.NewObjectPool[[]string](20)
	startLineArena := buffer.NewBuffer[byte](
		s.URL.BufferSize.Default,
		s.URL.BufferSize.Maximal,
	)
	parser := http1.NewHTTPRequestsParser(
		request, *keyArena, *valArena, *startLineArena, *objPool, s.Headers,
	)
	renderer := render.NewEngine(make([]byte, 0, 1024), nil, defaultHeaders)
	server := NewHTTPServer(r).(*httpServer)

	b.Run("simple POST", func(b *testing.B) {
		b.SetBytes(int64(len(simplePOST)))
		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			server.RunOnce(withBodyClient, request, renderer, parser)
		}
	})
}
