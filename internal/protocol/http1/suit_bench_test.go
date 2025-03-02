package http1

import (
	"github.com/indigo-web/indigo/config"
	"github.com/indigo-web/indigo/http"
	"github.com/indigo-web/indigo/internal/construct"
	"github.com/indigo-web/indigo/internal/requestgen"
	"github.com/indigo-web/indigo/router"
	"github.com/indigo-web/indigo/router/inbuilt"
	"github.com/indigo-web/indigo/transport"
	"github.com/indigo-web/indigo/transport/dummy"
	"strings"
	"testing"
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

	return r.Build()
}

func Benchmark_Get(b *testing.B) {
	b.Run("simple get", func(b *testing.B) {
		raw := []byte("GET / HTTP/1.1\r\nAccept-Encoding: identity\r\n")
		client := dummy.NewCircularClient(raw)
		server, _ := newSuit(client)
		b.SetBytes(int64(len(raw)))
		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			server.ServeOnce()
		}
	})

	b.Run("5 headers", func(b *testing.B) {
		data := requestgen.Generate(longPath, requestgen.Headers(5))
		client := dummy.NewCircularClient(data)
		server, _ := newSuit(client)
		b.SetBytes(int64(len(data)))
		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			server.ServeOnce()
		}
	})

	b.Run("10 headers", func(b *testing.B) {
		raw := requestgen.Generate(longPath, requestgen.Headers(10))
		dispersed := disperse(raw, config.Default().NET.ReadBufferSize)
		client := dummy.NewCircularClient(dispersed...)
		server, _ := newSuit(client)
		b.SetBytes(int64(len(raw)))
		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			for j := 0; j < len(dispersed); j++ {
				server.ServeOnce()
			}
		}
	})

	b.Run("50 headers", func(b *testing.B) {
		raw := requestgen.Generate(longPath, requestgen.Headers(50))
		dispersed := disperse(raw, config.Default().NET.ReadBufferSize)
		client := dummy.NewCircularClient(dispersed...)
		server, _ := newSuit(client)
		b.SetBytes(int64(len(raw)))
		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			for j := 0; j < len(dispersed); j++ {
				server.ServeOnce()
			}
		}
	})

	b.Run("heavily escaped", func(b *testing.B) {
		raw := requestgen.Generate(strings.Repeat("%20", 500), requestgen.Headers(10))
		dispersed := disperse(raw, config.Default().NET.ReadBufferSize)
		client := dummy.NewCircularClient(dispersed...)
		server, _ := newSuit(client)
		b.SetBytes(int64(len(raw)))
		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			for j := 0; j < len(dispersed); j++ {
				server.ServeOnce()
			}
		}
	})
}

func Benchmark_Post(b *testing.B) {
	b.Run("POST hello world", func(b *testing.B) {
		raw := []byte("POST / HTTP/1.1\r\nContent-Length: 13\r\n\r\nHello, world!")
		client := dummy.NewCircularClient(disperse(raw, config.Default().NET.ReadBufferSize)...)
		server, _ := newSuit(client)
		b.SetBytes(int64(len(raw)))
		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			server.ServeOnce()
		}
	})

	b.Run("discard POST 10mib", func(b *testing.B) {
		body := strings.Repeat("a", 10_000_000)
		raw := []byte("POST / HTTP/1.1\r\nContent-Length: 10000000\r\n\r\n" + body)
		client := dummy.NewCircularClient(disperse(raw, config.Default().NET.ReadBufferSize)...)
		server, _ := newSuit(client)
		b.SetBytes(int64(len(raw)))
		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			server.ServeOnce()
		}
	})

	b.Run("discard chunked 10mib", func(b *testing.B) {
		const chunkSize = 0xfffe
		const numberOfChunks = 10_000_000 / chunkSize
		chunk := "fffe\r\n" + strings.Repeat("a", chunkSize) + "\r\n"
		chunked := strings.Repeat(chunk, numberOfChunks) + "0\r\n\r\n"
		raw := []byte("POST / HTTP/1.1\r\nTransfer-Encoding: chunked\r\n\r\n" + chunked)
		client := dummy.NewCircularClient(disperse(raw, config.Default().NET.ReadBufferSize)...)
		server, _ := newSuit(client)
		b.SetBytes(int64(len(raw)))
		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			server.ServeOnce()
		}
	})
}

func newSuit(client transport.Client) (*Suit, *http.Request) {
	// using inbuilt router instead of simple in order to be more precise and realistic,
	// as in wildlife simple router will be barely used
	cfg := config.Default()
	r := getInbuiltRouter()
	body := NewBody(client, construct.Chunked(cfg.Body), cfg.Body)
	req := construct.Request(cfg, client, body)

	return Initialize(config.Default(), r, client, req, body), req
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
