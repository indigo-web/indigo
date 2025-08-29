package http1

import (
	"bytes"
	"fmt"
	"io"
	"net/http/httputil"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/indigo-web/indigo/config"
	"github.com/indigo-web/indigo/http"
	"github.com/indigo-web/indigo/http/codec"
	"github.com/indigo-web/indigo/http/method"
	"github.com/indigo-web/indigo/internal/codecutil"
	"github.com/indigo-web/indigo/internal/construct"
	"github.com/indigo-web/indigo/internal/httptest/serialize"
	"github.com/indigo-web/indigo/kv"
	"github.com/indigo-web/indigo/router"
	"github.com/indigo-web/indigo/router/inbuilt"
	"github.com/indigo-web/indigo/transport"
	"github.com/indigo-web/indigo/transport/dummy"
	"github.com/klauspost/compress/gzip"
	"github.com/stretchr/testify/require"
)

func generateHeaders(n int) http.Headers {
	hdrs := kv.NewPrealloc(n)

	for i := range n - 1 {
		hdrs.Add("some-random-header-name-nobody-cares-about"+strconv.Itoa(i), strings.Repeat("b", 50))
	}

	return hdrs.Add("Host", "localhost")
}

func generateRequest(uri string, hdrs http.Headers) (request []byte) {
	request = append(request, "GET /"+uri+" HTTP/1.1\r\n"...)

	for _, pair := range hdrs.Expose() {
		request = append(request, pair.Key+": "+pair.Value+"\r\n"...)
	}

	return append(request, "\r\n"...)
}

var longPath = strings.Repeat("a", 500)

func getInbuiltRouter() router.Router {
	r := inbuilt.New().
		Get("/"+longPath, http.Respond).
		Get("/with-header", func(request *http.Request) *http.Response {
			return request.Respond().Header("Hello", "World")
		}).
		Post("/echo", func(request *http.Request) *http.Response {
			return http.Stream(request, request.Body)
		})

	r.Resource("/").
		Get(http.Respond).
		Post(http.Respond)

	return r.Build()
}

func BenchmarkSuit(b *testing.B) {
	b.Run("GET root 5 headers", func(b *testing.B) {
		raw := generateRequest("", generateHeaders(5))
		client := dummy.NewMockClient(raw)
		server, request := getSuit(client)
		b.SetBytes(int64(len(raw)))
		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			server.ServeOnce()
			request.Reset()
		}
	})

	b.Run("GET long path 5 headers", func(b *testing.B) {
		data := generateRequest(longPath, generateHeaders(5))
		client := dummy.NewMockClient(data)
		server, request := getSuit(client)
		b.SetBytes(int64(len(data)))
		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			server.ServeOnce()
			request.Reset()
		}
	})

	b.Run("GET long path 10 headers", func(b *testing.B) {
		raw := generateRequest(longPath, generateHeaders(10))
		dispersed := scatter(raw, config.Default().NET.ReadBufferSize)
		client := dummy.NewMockClient(dispersed...)
		server, request := getSuit(client)
		b.SetBytes(int64(len(raw)))
		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			for j := 0; j < len(dispersed); j++ {
				server.ServeOnce()
				request.Reset()
			}
		}
	})

	b.Run("50 headers", func(b *testing.B) {
		raw := generateRequest(longPath, generateHeaders(50))
		dispersed := scatter(raw, config.Default().NET.ReadBufferSize)
		client := dummy.NewMockClient(dispersed...)
		server, request := getSuit(client)
		b.SetBytes(int64(len(raw)))
		b.ReportAllocs()
		b.ResetTimer()

		for range b.N {
			for j := 0; j < len(dispersed); j++ {
				server.ServeOnce()
				request.Reset()
			}
		}
	})

	b.Run("heavily escaped", func(b *testing.B) {
		raw := generateRequest(strings.Repeat("%20", 500), generateHeaders(10))
		dispersed := scatter(raw, config.Default().NET.ReadBufferSize)
		client := dummy.NewMockClient(dispersed...)
		server, request := getSuit(client)
		b.SetBytes(int64(len(raw)))
		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			for j := 0; j < len(dispersed); j++ {
				server.ServeOnce()
				request.Reset()
			}
		}
	})

	b.Run("POST hello world", func(b *testing.B) {
		raw := []byte("POST / HTTP/1.1\r\nContent-Length: 13\r\n\r\nHello, world!")
		client := dummy.NewMockClient(scatter(raw, config.Default().NET.ReadBufferSize)...)
		server, request := getSuit(client)
		b.SetBytes(int64(len(raw)))
		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			server.ServeOnce()
			request.Reset()
		}
	})
}

func getSuit(client transport.Client, codecs ...codec.Codec) (*Suit, *http.Request) {
	cfg := config.Default()
	r := getInbuiltRouter()
	req := construct.Request(cfg, client)
	suit := New(cfg, r, client, req, codecutil.NewCache(codecs, codecutil.AcceptEncoding(codecs)))
	req.Body = http.NewBody(suit)

	return suit, req
}

func scatter(data []byte, n int) (parts [][]byte) {
	for len(data) > 0 {
		boundary := min(n, len(data))
		parts = append(parts, data[:boundary])
		data = data[boundary:]
	}

	return parts
}

func encodeChunked(input []byte) []byte {
	out := new(bytes.Buffer)
	w := httputil.NewChunkedWriter(out)
	_, err := w.Write(input)
	if err != nil {
		panic(err)
	}

	err = w.Close()
	if err != nil {
		panic(err)
	}

	return out.Bytes()
}

func encodeGZIP(text string) []byte {
	buff := bytes.NewBuffer(nil)
	c := gzip.NewWriter(buff)
	_, err := c.Write([]byte(text))
	if err != nil {
		panic("unexpected error during gzipping")
	}
	if c.Close() != nil {
		panic("unexpected error during closing gzip writer")
	}

	return buff.Bytes()
}

func TestSuit(t *testing.T) {
	generateRequest := func(m method.Method, path string, headers http.Headers, body string) string {
		request := construct.Request(config.Default(), dummy.NewNopClient())
		request.Method = m
		request.Path = path
		request.Headers = headers.Add("Content-Length", strconv.Itoa(len(body)))

		request.TransferEncoding = slices.Collect(headers.Values("Transfer-Encoding"))
		request.ContentEncoding = slices.Collect(headers.Values("Content-Encoding"))

		return serialize.Headers(request) + body
	}

	t.Run("echo", func(t *testing.T) {
		const body = "Hello, world!"
		data := generateRequest(method.POST, "/echo", kv.New(), body)
		client := dummy.NewMockClient([]byte(data)).Journaling()
		server, _ := getSuit(client)
		require.True(t, server.ServeOnce())
		resp, err := parseHTTP11Response("POST", client.Written())
		require.NoError(t, err)
		require.Equal(t, 200, resp.StatusCode)
		b, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		require.Equal(t, body, string(b))
	})

	t.Run("decompress gzip", func(t *testing.T) {
		gzipped := encodeGZIP("Hello, world!")
		chunked := encodeChunked(gzipped)
		headers := kv.New().
			Add("Transfer-Encoding", "chunked").
			Add("Content-Encoding", "gzip")
		request := generateRequest(method.POST, "/echo", headers, string(chunked))
		client := dummy.NewMockClient([]byte(request)).Journaling()
		server, _ := getSuit(client, codec.NewGZIP())

		require.True(t, server.ServeOnce())
		resp, err := parseHTTP11Response("POST", client.Written())
		require.Equal(t, 200, resp.StatusCode)
		require.Equal(t, []string{"gzip"}, resp.Header["Accept-Encoding"])
		b, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		require.Equal(t, "Hello, world!", string(b))
	})
}

func TestPOST(t *testing.T) {
	// TODO: these test cases are unnecessary. They can be proven correct also operated on lesser data

	const N = 2

	t.Run("POST hello world", func(t *testing.T) {
		raw := []byte("POST / HTTP/1.1\r\nContent-Length: 13\r\n\r\nHello, world!")
		client := dummy.NewMockClient(scatter(raw, config.Default().NET.ReadBufferSize)...).LoopReads()
		server, _ := getSuit(client)

		for i := 0; i < N; i++ {
			require.True(t, server.ServeOnce())
		}
	})

	t.Run("discard POST 10mib", func(t *testing.T) {
		body := strings.Repeat("a", 10_000_000)
		raw := []byte("POST / HTTP/1.1\r\nContent-Length: 10000000\r\n\r\n" + body)
		dispersed := scatter(raw, config.Default().NET.ReadBufferSize)
		client := dummy.NewMockClient(dispersed...).LoopReads()
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
		dispersed := scatter(raw, config.Default().NET.ReadBufferSize)
		client := dummy.NewMockClient(dispersed...).LoopReads()
		server, _ := getSuit(client)

		for i := 0; i < N; i++ {
			for j := 0; j < len(dispersed); j++ {
				require.True(t, server.ServeOnce(), fmt.Sprintf("%d:%d", i, j))
			}
		}
	})
}
