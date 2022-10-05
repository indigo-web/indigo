package server

import (
	"context"
	"github.com/fakefloordiv/indigo/http/encodings"
	"github.com/fakefloordiv/indigo/http/headers"
	"github.com/fakefloordiv/indigo/http/parser/http1"
	render2 "github.com/fakefloordiv/indigo/http/render"
	"github.com/fakefloordiv/indigo/http/url"
	"github.com/fakefloordiv/indigo/router/inbuilt"
	"github.com/fakefloordiv/indigo/settings"
	"github.com/fakefloordiv/indigo/types"
	"net"
	"testing"
	"time"
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

type connMock struct {
	data []byte
}

func newConn(data []byte) net.Conn {
	return connMock{
		data: data,
	}
}

func (c connMock) Read(b []byte) (n int, err error) {
	copy(b, c.data)

	return len(b), nil
}

func (c connMock) Write(b []byte) (n int, err error) {
	return len(b), nil
}

func (c connMock) Close() error {
	return nil
}

func (c connMock) LocalAddr() net.Addr {
	return nil
}

func (c connMock) RemoteAddr() net.Addr {
	return nil
}

func (c connMock) SetDeadline(time.Time) error {
	return nil
}

func (c connMock) SetReadDeadline(time.Time) error {
	return nil
}

func (c connMock) SetWriteDeadline(time.Time) error {
	return nil
}

func BenchmarkIndigo(b *testing.B) {
	router := inbuilt.NewRouter()

	root := router.Resource("/")
	root.Get(func(context.Context, *types.Request) types.Response {
		return types.OK()
	})
	root.Post(func(_ context.Context, request *types.Request) types.Response {
		_ = request.OnBody(func([]byte) error {
			return nil
		}, func(error) {

		})

		return types.OK()
	})

	router.Get("/with-header", func(context.Context, *types.Request) types.Response {
		return types.WithHeader("Hello", "world")
	})

	router.Get("/with-two-headers", func(context.Context, *types.Request) types.Response {
		return types.
			WithHeader("Hello", "world").
			WithHeader("Lorem", "ipsum")
	})

	router.OnStart()

	s := settings.Default()
	manager := headers.NewManager(s.Headers)
	query := url.NewQuery(func() map[string][]byte {
		return make(map[string][]byte)
	})
	request, writer := types.NewRequest(&manager, query, nil)
	startLineBuff := make([]byte, 0, s.URL.Length.Maximal)
	headerBuff := make([]byte, 0, s.Headers.KeyLength.Maximal)
	codings := encodings.NewContentEncodings()
	parser := http1.NewHTTPRequestsParser(request, writer, startLineBuff, headerBuff, s, &manager, codings)

	// because only tcp server reads from conn. We do not benchmark tcp server here
	conn := newConn(nil)
	defaultHeaders := make(headers.Headers, 10)
	render := render2.NewRenderer(make([]byte, 0, 1024), defaultHeaders)

	server := NewHTTPServer(request, router, parser, conn, render)
	go server.Run()

	b.Run("SimpleGET", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = server.OnData(simpleGETRequest)
		}
	})

	b.Run("FiveHeadersGET", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = server.OnData(fiveHeadersGETRequest)
		}
	})

	b.Run("TenHeadersGET", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = server.OnData(tenHeadersGETRequest)
		}
	})

	b.Run("SimpleGETWithHeader", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = server.OnData(simpleGETWithHeader)
		}
	})

	b.Run("SimplePOST", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = server.OnData(simplePOST)
		}
	})
}
