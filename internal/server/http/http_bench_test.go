package http

import (
	"strings"
	"testing"

	"github.com/indigo-web/indigo/http"
	"github.com/indigo-web/indigo/http/coding"
	"github.com/indigo-web/indigo/http/headers"
	"github.com/indigo-web/indigo/http/method"
	"github.com/indigo-web/indigo/http/query"
	"github.com/indigo-web/indigo/http/status"
	"github.com/indigo-web/indigo/internal/requestgen"
	"github.com/indigo-web/indigo/internal/server/tcp/dummy"
	"github.com/indigo-web/indigo/internal/transport"
	"github.com/indigo-web/indigo/internal/transport/http1"
	"github.com/indigo-web/indigo/router"
	"github.com/indigo-web/indigo/router/inbuilt"
	"github.com/indigo-web/indigo/router/simple"
	"github.com/indigo-web/indigo/settings"
	"github.com/indigo-web/utils/buffer"
	"github.com/indigo-web/utils/pool"
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
			return request.Respond().Header("Hello", "World")
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
				return http.Error(request, status.ErrMethodNotAllowed)
			}
		case "/with-header":
			return request.Respond().Header("Hello", "World")
		case longpath:
			return request.Respond()
		default:
			return http.Error(request, status.ErrNotFound)
		}
	}, http.Error)

	return r
}

func Benchmark_Get(b *testing.B) {
	server, request, trans := newServer()

	b.Run("simple get", func(b *testing.B) {
		tenHeadersGETClient := dummy.NewCircularClient(tenHeadersGETRequest)
		b.SetBytes(int64(len(tenHeadersGETRequest)))
		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			server.HandleRequest(tenHeadersGETClient, request, trans)
		}
	})

	b.Run("with resp header", func(b *testing.B) {
		withRespHeadersGETClient := dummy.NewCircularClient(simpleGETWithHeader)
		b.SetBytes(int64(len(simpleGETWithHeader)))
		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			server.HandleRequest(withRespHeadersGETClient, request, trans)
		}
	})

	b.Run("5 headers", func(b *testing.B) {
		data := requestgen.Generate(longPath, 5)
		client := dummy.NewCircularClient(data)
		b.SetBytes(int64(len(data)))
		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			server.HandleRequest(client, request, trans)
		}
	})

	b.Run("10 headers", func(b *testing.B) {
		data := requestgen.Generate(longPath, 10)
		client := dummy.NewCircularClient(data)
		b.SetBytes(int64(len(data)))
		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			server.HandleRequest(client, request, trans)
		}
	})

	b.Run("50 headers", func(b *testing.B) {
		data := requestgen.Generate(longPath, 50)
		client := dummy.NewCircularClient(data)
		b.SetBytes(int64(len(data)))
		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			server.HandleRequest(client, request, trans)
		}
	})

	b.Run("heavily escaped", func(b *testing.B) {
		data := requestgen.Generate(strings.Repeat("%20", 500), 20)
		client := dummy.NewCircularClient(data)
		b.SetBytes(int64(len(data)))
		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			server.HandleRequest(client, request, trans)
		}
	})
}

func Benchmark_Post(b *testing.B) {
	withBodyClient := dummy.NewCircularClient(simplePOST)
	server, request, trans := newServer()

	b.Run("simple POST", func(b *testing.B) {
		b.SetBytes(int64(len(simplePOST)))
		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			server.HandleRequest(withBodyClient, request, trans)
		}
	})
}

func newServer() (*Server, *http.Request, transport.Transport) {
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
		hdrs, q, http.NewResponse(), dummy.NewNopConn(), body, nil, false,
	)
	keyBuff := buffer.NewBuffer[byte](
		s.Headers.MaxKeyLength*s.Headers.Number.Default,
		s.Headers.MaxKeyLength*s.Headers.Number.Maximal,
	)
	valBuff := buffer.NewBuffer[byte](
		s.Headers.ValueSpace.Default, s.Headers.ValueSpace.Maximal,
	)
	objPool := pool.NewObjectPool[[]string](20)
	startLineBuff := buffer.NewBuffer[byte](
		s.URL.BufferSize.Default,
		s.URL.BufferSize.Maximal,
	)
	trans := http1.New(
		request,
		*keyBuff, *valBuff, *startLineBuff,
		*objPool,
		s.Headers,
		make([]byte, 0, 1024),
		nil,
		defaultHeaders,
	)

	return NewServer(r), request, trans
}
