package http

import (
	"github.com/indigo-web/indigo/internal/initialize"
	"github.com/indigo-web/indigo/internal/server/tcp"
	"strings"
	"testing"

	"github.com/indigo-web/indigo/http"
	"github.com/indigo-web/indigo/internal/requestgen"
	"github.com/indigo-web/indigo/internal/server/tcp/dummy"
	"github.com/indigo-web/indigo/internal/transport"
	"github.com/indigo-web/indigo/router"
	"github.com/indigo-web/indigo/router/inbuilt"
	"github.com/indigo-web/indigo/settings"
)

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
			_ = request.Body.Callback(func(b []byte) error {
				return nil
			})

			return request.Respond()
		})

	if err := r.OnStart(); err != nil {
		panic(err)
	}

	return r
}

func Benchmark_Get(b *testing.B) {
	b.Run("simple get", func(b *testing.B) {
		server, request, trans := newServer(dummy.NewNopClient())
		raw := []byte("GET / HTTP/1.1\r\nAccept-Encoding: identity\r\n")
		client := dummy.NewCircularClient(raw)
		b.SetBytes(int64(len(raw)))
		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			server.HandleRequest(client, request, trans)
		}
	})

	b.Run("5 headers", func(b *testing.B) {
		server, request, trans := newServer(dummy.NewNopClient())
		data := requestgen.Generate(longPath, requestgen.Headers(5))
		client := dummy.NewCircularClient(data)
		b.SetBytes(int64(len(data)))
		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			server.HandleRequest(client, request, trans)
		}
	})

	b.Run("10 headers", func(b *testing.B) {
		server, request, trans := newServer(dummy.NewNopClient())
		raw := requestgen.Generate(longPath, requestgen.Headers(10))
		dispersed := disperse(raw, settings.Default().TCP.ReadBufferSize)
		client := dummy.NewCircularClient(dispersed...)
		b.SetBytes(int64(len(raw)))
		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			for j := 0; j < len(dispersed); j++ {
				server.HandleRequest(client, request, trans)
			}
		}
	})

	b.Run("50 headers", func(b *testing.B) {
		server, request, trans := newServer(dummy.NewNopClient())
		raw := requestgen.Generate(longPath, requestgen.Headers(50))
		dispersed := disperse(raw, settings.Default().TCP.ReadBufferSize)
		client := dummy.NewCircularClient(dispersed...)
		b.SetBytes(int64(len(raw)))
		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			for j := 0; j < len(dispersed); j++ {
				server.HandleRequest(client, request, trans)
			}
		}
	})

	b.Run("heavily escaped", func(b *testing.B) {
		server, request, trans := newServer(dummy.NewNopClient())
		raw := requestgen.Generate(strings.Repeat("%20", 500), requestgen.Headers(10))
		dispersed := disperse(raw, settings.Default().TCP.ReadBufferSize)
		client := dummy.NewCircularClient(dispersed...)
		b.SetBytes(int64(len(raw)))
		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			for j := 0; j < len(dispersed); j++ {
				server.HandleRequest(client, request, trans)
			}
		}
	})
}

func Benchmark_Post(b *testing.B) {
	b.Run("POST hello world", func(b *testing.B) {
		raw := []byte("POST / HTTP/1.1\r\nContent-Length: 13\r\n\r\nHello, world!")
		client := dummy.NewCircularClient(disperse(raw, settings.Default().TCP.ReadBufferSize)...)
		server, request, trans := newServer(client)
		b.SetBytes(int64(len(raw)))
		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			server.HandleRequest(client, request, trans)
		}
	})

	b.Run("discard POST 10mib", func(b *testing.B) {
		body := strings.Repeat("a", 10_000_000)
		raw := []byte("POST / HTTP/1.1\r\nContent-Length: 10000000\r\n\r\n" + body)
		client := dummy.NewCircularClient(disperse(raw, settings.Default().TCP.ReadBufferSize)...)
		server, request, trans := newServer(client)
		b.SetBytes(int64(len(raw)))
		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			server.HandleRequest(client, request, trans)
		}
	})

	b.Run("discard chunked 10mib", func(b *testing.B) {
		const chunkSize = 0xfffe
		const numberOfChunks = 10_000_000 / chunkSize
		chunk := "fffe\r\n" + strings.Repeat("a", chunkSize) + "\r\n"
		chunked := strings.Repeat(chunk, numberOfChunks) + "0\r\n\r\n"
		raw := []byte("POST / HTTP/1.1\r\nTransfer-Encoding: chunked\r\n\r\n" + chunked)
		client := dummy.NewCircularClient(disperse(raw, settings.Default().TCP.ReadBufferSize)...)
		server, request, trans := newServer(client)
		b.SetBytes(int64(len(raw)))
		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			server.HandleRequest(client, request, trans)
		}
	})
}

func newServer(client tcp.Client) (*Server, *http.Request, transport.Transport) {
	// using inbuilt router instead of simple in order to be more precise and realistic,
	// as in wildlife simple router will be barely used
	r := getInbuiltRouter()
	body := initialize.NewBody(client, settings.Default().Body)
	request := initialize.NewRequest(settings.Default(), dummy.NewNopConn(), body)
	trans := initialize.NewTransport(settings.Default(), request)

	return NewServer(r, nil), request, trans
}

func disperse(data []byte, n int) (parts [][]byte) {
	for len(data) > 0 {
		end := min(len(data), n)
		part := data[:end]
		parts = append(parts, part)
		data = data[end:]
	}

	return parts
}

func min(a, b int) int {
	if a < b {
		return a
	}

	return b
}
