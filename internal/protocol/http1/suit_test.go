package http1

import (
	"fmt"
	"github.com/indigo-web/indigo/config"
	"github.com/indigo-web/indigo/http"
	"github.com/indigo-web/indigo/internal/codecutil"
	"github.com/indigo-web/indigo/internal/construct"
	"github.com/indigo-web/indigo/kv"
	"github.com/indigo-web/indigo/router"
	"github.com/indigo-web/indigo/router/inbuilt"
	"github.com/indigo-web/indigo/router/simple"
	"github.com/indigo-web/indigo/transport"
	"github.com/indigo-web/indigo/transport/dummy"
	"github.com/stretchr/testify/require"
	"strconv"
	"strings"
	"testing"
)

func newSimpleRouter(t *testing.T, want http.Headers) *simple.Router {
	return simple.New(func(request *http.Request) *http.Response {
		require.True(t, compareHeaders(want, request.Headers))
		return http.Respond(request)
	}, func(request *http.Request) *http.Response {
		require.Failf(t, "unexpected error", "unexpected error: %s", request.Env.Error.Error())
		return nil
	})
}

func GenerateHeaders(n int) http.Headers {
	hdrs := kv.NewPrealloc(n)

	for i := 0; i < n-1; i++ {
		hdrs.Add("some-random-header-name-nobody-cares-about"+strconv.Itoa(i), strings.Repeat("b", 100))
	}

	return hdrs.Add("Host", "localhost")
}

func HeadersBlock(hdrs http.Headers) (buff []byte) {
	for _, pair := range hdrs.Expose() {
		buff = append(buff, pair.Key+": "+pair.Value+"\r\n"...)
	}

	return buff
}

func GenerateRequest(uri string, hdrs http.Headers) (request []byte) {
	request = append(request, "GET /"+uri+" HTTP/1.1\r\n"...)
	request = append(request, HeadersBlock(hdrs)...)

	return append(request, '\r', '\n')
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
		client := dummy.NewClient(raw)
		server, _ := getSuit(client)
		b.SetBytes(int64(len(raw)))
		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			server.ServeOnce()
		}
	})

	b.Run("5 headers", func(b *testing.B) {
		data := GenerateRequest(longPath, GenerateHeaders(5))
		client := dummy.NewClient(data)
		server, _ := getSuit(client)
		b.SetBytes(int64(len(data)))
		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			server.ServeOnce()
		}
	})

	b.Run("10 headers", func(b *testing.B) {
		raw := GenerateRequest(longPath, GenerateHeaders(10))
		dispersed := disperse(raw, config.Default().NET.ReadBufferSize)
		client := dummy.NewClient(dispersed...)
		server, _ := getSuit(client)
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
		raw := GenerateRequest(longPath, GenerateHeaders(50))
		dispersed := disperse(raw, config.Default().NET.ReadBufferSize)
		client := dummy.NewClient(dispersed...)
		server, _ := getSuit(client)
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
		raw := GenerateRequest(strings.Repeat("%20", 500), GenerateHeaders(10))
		dispersed := disperse(raw, config.Default().NET.ReadBufferSize)
		client := dummy.NewClient(dispersed...)
		server, _ := getSuit(client)
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
		client := dummy.NewClient(disperse(raw, config.Default().NET.ReadBufferSize)...)
		server, _ := getSuit(client)
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
		client := dummy.NewClient(disperse(raw, config.Default().NET.ReadBufferSize)...)
		server, _ := getSuit(client)
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
		client := dummy.NewClient(disperse(raw, config.Default().NET.ReadBufferSize)...)
		server, _ := getSuit(client)
		b.SetBytes(int64(len(raw)))
		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			server.ServeOnce()
		}
	})
}

func getSuit(client transport.Client) (*Suit, *http.Request) {
	// using inbuilt router instead of simple in order to be more precise and realistic,
	// as in wildlife simple router will be barely used
	cfg := config.Default()
	r := getInbuiltRouter()
	req := construct.Request(cfg, client)
	suit := New(cfg, r, client, req, codecutil.Cache[http.Decompressor]{})
	req.Body = http.NewBody(cfg, suit)

	return suit, req
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

func TestServer(t *testing.T) {
	const N = 10

	t.Run("simple get", func(t *testing.T) {
		raw := []byte("GET / HTTP/1.1\r\nAccept-Encoding: identity\r\n\r\n")
		client := dummy.NewCircularClient(raw)
		server, _ := getSuit(client)
		wantHeaders := kv.New().Add("Accept-Encoding", "identity")
		server.router = newSimpleRouter(t, wantHeaders)

		for i := 0; i < N; i++ {
			require.True(t, server.ServeOnce())
		}
	})

	t.Run("5 headers", func(t *testing.T) {
		wantHeaders := GenerateHeaders(5)
		raw := GenerateRequest(longPath, wantHeaders)
		dispersed := disperse(raw, config.Default().NET.ReadBufferSize)
		client := dummy.NewCircularClient(dispersed...)
		server, _ := getSuit(client)
		server.router = newSimpleRouter(t, wantHeaders)

		for i := 0; i < N; i++ {
			require.True(t, server.ServeOnce())
		}
	})

	t.Run("10 headers", func(t *testing.T) {
		wantHeaders := GenerateHeaders(10)
		raw := GenerateRequest(longPath, wantHeaders)
		dispersed := disperse(raw, config.Default().NET.ReadBufferSize)
		client := dummy.NewCircularClient(dispersed...)
		server, _ := getSuit(client)
		server.router = newSimpleRouter(t, wantHeaders)

		for i := 0; i < N; i++ {
			require.True(t, server.ServeOnce())
		}
	})

	t.Run("50 headers", func(t *testing.T) {
		wantHeaders := GenerateHeaders(50)
		raw := GenerateRequest(longPath, wantHeaders)
		dispersed := disperse(raw, config.Default().NET.ReadBufferSize)
		client := dummy.NewCircularClient(dispersed...)
		server, _ := getSuit(client)
		server.router = newSimpleRouter(t, wantHeaders)

		for i := 0; i < N; i++ {
			for j := 0; j < len(dispersed); j++ {
				require.True(t, server.ServeOnce())
			}
		}
	})

	t.Run("heavily escaped", func(t *testing.T) {
		wantHeaders := GenerateHeaders(20)
		raw := GenerateRequest(strings.Repeat("%20", 500), wantHeaders)
		dispersed := disperse(raw, config.Default().NET.ReadBufferSize)
		client := dummy.NewCircularClient(dispersed...)
		server, _ := getSuit(client)
		server.router = newSimpleRouter(t, wantHeaders)

		for i := 0; i < N; i++ {
			for j := 0; j < len(dispersed); j++ {
				require.True(t, server.ServeOnce())
			}
		}
	})
}

func TestPOST(t *testing.T) {
	const N = 10

	t.Run("POST hello world", func(t *testing.T) {
		raw := []byte("POST / HTTP/1.1\r\nContent-Length: 13\r\n\r\nHello, world!")
		client := dummy.NewClient(disperse(raw, config.Default().NET.ReadBufferSize)...)
		server, _ := getSuit(client)

		for i := 0; i < N; i++ {
			require.True(t, server.ServeOnce())
		}
	})

	t.Run("discard POST 10mib", func(t *testing.T) {
		body := strings.Repeat("a", 10_000_000)
		raw := []byte("POST / HTTP/1.1\r\nContent-Length: 10000000\r\n\r\n" + body)
		dispersed := disperse(raw, config.Default().NET.ReadBufferSize)
		client := dummy.NewClient(dispersed...)
		server, _ := getSuit(client)

		for i := 0; i < N; i++ {
			for j := 0; j < len(dispersed); j++ {
				require.True(t, server.ServeOnce())
			}
		}
	})

	t.Run("discard chunked 10mib", func(t *testing.T) {
		const chunkSize = 0xfffe
		const numberOfChunks = 10_000_000 / chunkSize
		chunk := "fffe\r\n" + strings.Repeat("a", chunkSize) + "\r\n"
		chunked := strings.Repeat(chunk, numberOfChunks) + "0\r\n\r\n"
		raw := []byte("POST / HTTP/1.1\r\nTransfer-Encoding: chunked\r\n\r\n" + chunked)
		dispersed := disperse(raw, config.Default().NET.ReadBufferSize)
		client := dummy.NewClient(dispersed...)
		server, _ := getSuit(client)

		for i := 0; i < N; i++ {
			for j := 0; j < len(dispersed); j++ {
				require.True(t, server.ServeOnce(), fmt.Sprintf("%d:%d", i, j))
			}
		}
	})
}

func compareHeaders(a, b http.Headers) bool {
	first, second := a.Expose(), b.Expose()
	if len(first) != len(second) {
		return false
	}

	for i, pair := range first {
		if pair != second[i] {
			return false
		}
	}

	return true
}
