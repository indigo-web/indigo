package http1

import (
	"errors"
	"fmt"
	"github.com/dchest/uniuri"
	"github.com/indigo-web/indigo/config"
	"github.com/indigo-web/indigo/http"
	"github.com/indigo-web/indigo/http/method"
	"github.com/indigo-web/indigo/http/proto"
	"github.com/indigo-web/indigo/http/status"
	"github.com/indigo-web/indigo/internal/construct"
	"github.com/indigo-web/indigo/kv"
	"github.com/indigo-web/indigo/transport/dummy"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
)

func getParser(cfg *config.Config) (*parser, *http.Request) {
	request := construct.Request(cfg, dummy.NewNopClient())
	keys, values, requestLine := construct.Buffers(cfg)
	p := newParser(cfg, request, keys, values, requestLine)

	return p, request
}

func BenchmarkParser(b *testing.B) {
	parser, request := getParser(config.Default())

	b.Run("with 5 headers", func(b *testing.B) {
		data := GenerateRequest(strings.Repeat("a", 500), GenerateHeaders(5))
		b.SetBytes(int64(len(data)))
		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			_, _, _ = parser.Parse(data)
			request.Reset()
		}
	})

	b.Run("with 10 headers", func(b *testing.B) {
		data := GenerateRequest(strings.Repeat("a", 500), GenerateHeaders(10))
		b.SetBytes(int64(len(data)))
		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			_, _, _ = parser.Parse(data)
			request.Reset()
		}
	})

	b.Run("with 50 headers", func(b *testing.B) {
		data := GenerateRequest(strings.Repeat("a", 500), GenerateHeaders(50))
		b.SetBytes(int64(len(data)))
		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			_, _, _ = parser.Parse(data)
			request.Reset()
		}
	})

	b.Run("escaped 10 headers", func(b *testing.B) {
		data := GenerateRequest(strings.Repeat("%20", 500), GenerateHeaders(10))
		b.SetBytes(int64(len(data)))
		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			_, _, _ = parser.Parse(data)
			request.Reset()
		}
	})
}

type wantedRequest struct {
	Headers  http.Headers
	Path     string
	Method   method.Method
	Protocol proto.Protocol
}

func compareRequests(t *testing.T, wanted wantedRequest, actual *http.Request) {
	require.Equal(t, wanted.Method, actual.Method)
	require.Equal(t, wanted.Path, actual.Path)
	require.Equal(t, wanted.Protocol, actual.Protocol)

	keys := wanted.Headers.Keys()

	for _, key := range keys {
		require.Equal(t, wanted.Headers.Values(key), actual.Headers.Values(key))
	}
}

func splitIntoParts(req []byte, n int) (parts [][]byte) {
	for i := 0; i < len(req); i += n {
		end := i + n
		if end > len(req) {
			end = len(req)
		}

		parts = append(parts, req[i:end])
	}

	return parts
}

func feedPartially(p *parser, raw []byte, n int) (done bool, extra []byte, err error) {
	parts := splitIntoParts(raw, n)

	for i, chunk := range parts {
		done, extra, err = p.Parse(chunk)
		if err != nil {
			return done, extra, err
		}
		if done {
			if i+1 < len(parts) {
				return true, extra, errors.New("not all chunks were fed")
			}

			break
		}
	}

	return done, extra, err
}

func TestParser(t *testing.T) {
	cfg := config.Default()
	cfg.Headers.MaxEncodingTokens = 3

	t.Run("simple GET", func(t *testing.T) {
		raw := "GET / HTTP/1.1\r\n\r\n"
		parser, request := getParser(cfg)
		done, extra, err := parser.Parse([]byte(raw))
		require.NoError(t, err)
		require.True(t, done)
		require.Empty(t, extra)

		wanted := wantedRequest{
			Method:   method.GET,
			Path:     "/",
			Protocol: proto.HTTP11,
			Headers:  kv.New(),
		}

		compareRequests(t, wanted, request)
		request.Reset()
	})

	t.Run("simple GET with leading CRLF", func(t *testing.T) {
		raw := "\r\n\r\nGET / HTTP/1.1\r\n\r\n"
		parser, request := getParser(cfg)
		done, extra, err := parser.Parse([]byte(raw))
		require.Error(t, err, status.ErrBadRequest.Error())
		require.True(t, done)
		require.Empty(t, extra)
		request.Reset()
	})

	t.Run("normal GET", func(t *testing.T) {
		raw := "GET / HTTP/1.1\r\nHello: World!\r\nEaster: Egg\r\n\r\n"
		parser, request := getParser(cfg)
		done, extra, err := parser.Parse([]byte(raw))
		require.NoError(t, err)
		require.True(t, done)
		require.Empty(t, extra)

		wanted := wantedRequest{
			Method:   method.GET,
			Path:     "/",
			Protocol: proto.HTTP11,
			Headers: kv.NewFromMap(map[string][]string{
				"hello": {"World!"},
			}),
		}

		compareRequests(t, wanted, request)
		request.Reset()
	})

	t.Run("multiple header values", func(t *testing.T) {
		raw := "GET / HTTP/1.1\r\nAccept: one,two\r\nAccept: three\r\n\r\n"
		parser, request := getParser(cfg)
		done, extra, err := parser.Parse([]byte(raw))
		require.NoError(t, err)
		require.True(t, done)
		require.Empty(t, extra)

		wanted := wantedRequest{
			Method:   method.GET,
			Path:     "/",
			Protocol: proto.HTTP11,
			Headers: kv.NewFromMap(map[string][]string{
				"accept": {"one,two", "three"},
			}),
		}

		compareRequests(t, wanted, request)
		request.Reset()
	})

	t.Run("only lf", func(t *testing.T) {
		raw := "GET / HTTP/1.1\nHello: World!\n\n"
		parser, request := getParser(cfg)
		done, extra, err := parser.Parse([]byte(raw))
		require.NoError(t, err)
		require.True(t, done)
		require.Empty(t, extra)

		wanted := wantedRequest{
			Method:   method.GET,
			Path:     "/",
			Protocol: proto.HTTP11,
			Headers: kv.NewFromMap(map[string][]string{
				"hello": {"World!"},
			}),
		}

		compareRequests(t, wanted, request)
		request.Reset()
	})

	t.Run("escaping", func(t *testing.T) {
		parser, request := getParser(cfg)
		tcs := []struct {
			Raw      []byte
			WantPath string
		}{
			{
				[]byte("GET /hello%2C%20world HTTP/1.1\r\n\r\n"),
				"/hello, world",
			},
			{
				GenerateRequest(strings.Repeat("%20", 500), GenerateHeaders(10)),
				"/" + strings.Repeat(" ", 500),
			},
		}

		for i, tc := range tcs {
			done, extra, err := parser.Parse(tc.Raw)
			require.NoError(t, err, i)
			require.True(t, done, i)
			require.Empty(t, extra, i)

			wanted := wantedRequest{
				Method:   method.GET,
				Path:     tc.WantPath,
				Protocol: proto.HTTP11,
				Headers:  kv.New(),
			}

			compareRequests(t, wanted, request)
			request.Reset()
		}

	})

	t.Run("fuzz GET", func(t *testing.T) {
		raw := "GET / HTTP/1.1\r\nHello: World!\r\nEaster: Egg\r\n\r\n"
		parser, request := getParser(cfg)

		for i := 1; i < len(raw); i++ {
			done, extra, err := feedPartially(parser, []byte(raw), i)
			require.NoError(t, err, i)
			require.Empty(t, extra)
			require.True(t, done)

			wanted := wantedRequest{
				Method:   method.GET,
				Path:     "/",
				Protocol: proto.HTTP11,
				Headers: kv.NewFromMap(map[string][]string{
					"hello": {"World!"},
				}),
			}

			compareRequests(t, wanted, request)
			request.Reset()
		}
	})

	t.Run("absolute path", func(t *testing.T) {
		raw := "GET http://www.w3.org/pub/WWW/TheProject.html HTTP/1.1\r\n\r\n"
		parser, request := getParser(cfg)
		done, extra, err := parser.Parse([]byte(raw))
		require.NoError(t, err)
		require.True(t, done)
		require.Empty(t, extra)

		wanted := wantedRequest{
			Method:   method.GET,
			Path:     "http://www.w3.org/pub/WWW/TheProject.html",
			Protocol: proto.HTTP11,
			Headers:  kv.New(),
		}

		compareRequests(t, wanted, request)
		request.Reset()
	})

	t.Run("content-length", func(t *testing.T) {
		raw := "GET / HTTP/1.1\r\nContent-Length: 13\n\r\nHello, world!"
		parser, request := getParser(cfg)
		done, extra, err := parser.Parse([]byte(raw))
		require.NoError(t, err)
		require.True(t, done)
		require.Equal(t, "Hello, world!", string(extra))
		require.Equal(t, 13, request.ContentLength)
		request.Reset()

		raw = "GET / HTTP/1.1\r\nContent-Length: 13\r\nHi-Hi: ha-ha\r\n\r\nHello, world!"
		done, extra, err = parser.Parse([]byte(raw))
		require.NoError(t, err)
		require.True(t, done)
		require.Equal(t, "Hello, world!", string(extra))
		require.Equal(t, 13, request.ContentLength)
		require.True(t, request.Headers.Has("hi-hi"))
		require.Equal(t, "ha-ha", request.Headers.Value("hi-hi"))
		request.Reset()
	})

	t.Run("connection", func(t *testing.T) {
		raw := "GET / HTTP/1.1\r\nConnection: Keep-Alive\r\n\r\n"
		parser, request := getParser(cfg)
		done, extra, err := parser.Parse([]byte(raw))
		require.NoError(t, err)
		require.True(t, done)
		require.Empty(t, string(extra))
		require.Equal(t, "Keep-Alive", request.Connection)
		request.Reset()
	})

	t.Run("Transfer-Encoding and Content-Encoding", func(t *testing.T) {
		raw := "GET / HTTP/1.1\r\nTransfer-Encoding: chunked\r\nContent-Encoding: gzip, deflate\r\n\r\n"
		parser, request := getParser(cfg)
		done, extra, err := parser.Parse([]byte(raw))
		require.NoError(t, err)
		require.True(t, done)
		require.Empty(t, string(extra))
		require.Equal(t, []string{"chunked"}, request.Encoding.Transfer)
		require.True(t, request.Encoding.Chunked)
		require.Equal(t, []string{"gzip", "deflate"}, request.Encoding.Content)
		request.Reset()
	})

	t.Run("path params", func(t *testing.T) {
		raw := "GET /path?hello=world HTTP/1.1\r\n\r\n"
		parser, request := getParser(cfg)
		done, extra, err := parser.Parse([]byte(raw))
		require.NoError(t, err)
		require.True(t, done)
		require.Empty(t, extra)

		wanted := wantedRequest{
			Method:   method.GET,
			Path:     "/path",
			Protocol: proto.HTTP11,
			Headers:  kv.New(),
		}

		compareRequests(t, wanted, request)
		require.Equal(t, "world", request.Params.Value("hello"))
		request.Reset()
	})
}

func TestParser_Edgecases(t *testing.T) {
	t.Run("no method", func(t *testing.T) {
		parser, _ := getParser(config.Default())
		raw := []byte(" / HTTP/1.1\r\n\r\n")
		done, _, err := parser.Parse(raw)
		require.EqualError(t, err, status.ErrBadRequest.Error())
		require.True(t, done)
	})

	t.Run("no path", func(t *testing.T) {
		parser, _ := getParser(config.Default())
		raw := []byte("GET HTTP/1.1\r\n\r\n")
		done, _, err := parser.Parse(raw)
		require.EqualError(t, err, status.ErrBadRequest.Error())
		require.True(t, done)
	})

	t.Run("whitespace as a path", func(t *testing.T) {
		parser, _ := getParser(config.Default())
		raw := []byte("GET  HTTP/1.1\r\n\r\n")
		done, _, err := parser.Parse(raw)
		require.EqualError(t, err, status.ErrBadRequest.Error())
		require.True(t, done)
	})

	t.Run("short invalid method", func(t *testing.T) {
		parser, _ := getParser(config.Default())
		raw := []byte("GE / HTTP/1.1\r\n\r\n")
		done, _, err := parser.Parse(raw)
		require.EqualError(t, err, status.ErrMethodNotImplemented.Error())
		require.True(t, done)
	})

	t.Run("normal invalid method", func(t *testing.T) {
		parser, _ := getParser(config.Default())
		raw := []byte("GOT / HTTP/1.1\r\n\r\n")
		done, _, err := parser.Parse(raw)
		require.EqualError(t, err, status.ErrMethodNotImplemented.Error())
		require.True(t, done)
	})

	t.Run("long invalid method", func(t *testing.T) {
		parser, _ := getParser(config.Default())
		raw := []byte("PATCHPOSTPUT / HTTP/1.1\r\n\r\n")
		done, _, err := parser.Parse(raw)
		require.EqualError(t, err, status.ErrMethodNotImplemented.Error())
		require.True(t, done)
	})

	t.Run("short invalid protocol", func(t *testing.T) {
		parser, _ := getParser(config.Default())
		raw := []byte("GET / HTT\r\n\r\n")
		done, _, err := parser.Parse(raw)
		require.EqualError(t, err, status.ErrHTTPVersionNotSupported.Error())
		require.True(t, done)
	})

	t.Run("long invalid protocol", func(t *testing.T) {
		parser, _ := getParser(config.Default())
		raw := []byte("GET / HTTPS/1.1\r\n\r\n")
		done, _, err := parser.Parse(raw)
		require.EqualError(t, err, status.ErrHTTPVersionNotSupported.Error())
		require.True(t, done)
	})

	t.Run("unsupported minor version", func(t *testing.T) {
		parser, _ := getParser(config.Default())
		raw := []byte("GET / HTTP/1.2\r\n\r\n")
		done, _, err := parser.Parse(raw)
		require.EqualError(t, err, status.ErrHTTPVersionNotSupported.Error())
		require.True(t, done)
	})

	t.Run("unsupported major version", func(t *testing.T) {
		parser, _ := getParser(config.Default())
		raw := []byte("GET / HTTP/42.1\r\n\r\n")
		done, _, err := parser.Parse(raw)
		require.EqualError(t, err, status.ErrHTTPVersionNotSupported.Error())
		require.True(t, done)
	})

	t.Run("invalid minor version", func(t *testing.T) {
		parser, _ := getParser(config.Default())
		raw := []byte("GET / HTTP/1.x\r\n\r\n")
		done, _, err := parser.Parse(raw)
		require.EqualError(t, err, status.ErrHTTPVersionNotSupported.Error())
		require.True(t, done)
	})

	t.Run("invalid major version", func(t *testing.T) {
		parser, _ := getParser(config.Default())
		raw := []byte("GET / HTTP/x.1\r\n\r\n")
		done, _, err := parser.Parse(raw)
		require.EqualError(t, err, status.ErrHTTPVersionNotSupported.Error())
		require.True(t, done)
	})

	t.Run("lfcr crlf break sequence", func(t *testing.T) {
		parser, _ := getParser(config.Default())
		raw := []byte("GET / HTTP/1.1\n\r\r\n")
		done, _, err := parser.Parse(raw)
		require.EqualError(t, err, status.ErrBadRequest.Error())
		require.True(t, done)
	})

	t.Run("lfcr lfcr break sequence", func(t *testing.T) {
		// our parser is able to parse both crlf and lf splitters
		// so in example below he sees LF CRLF CR
		// the last one CR will be returned as extra-bytes
		parser, _ := getParser(config.Default())
		raw := []byte("GET / HTTP/1.1\n\r\n\r")
		done, extra, err := parser.Parse(raw)
		require.Equal(t, []byte("\r"), extra)
		require.NoError(t, err)
		require.True(t, done)
	})

	t.Run("invalid content length", func(t *testing.T) {
		parser, _ := getParser(config.Default())
		raw := []byte("GET / HTTP/1.1\r\nContent-Length: 1f5\r\n\r\n")
		_, _, err := parser.Parse(raw)
		require.EqualError(t, err, status.ErrBadRequest.Error())
	})

	t.Run("simple request", func(t *testing.T) {
		// Simple Requests are not supported, because our server is
		// HTTP/1.1-oriented, and in 1.1 simple request/response is
		// something like a deprecated mechanism
		parser, _ := getParser(config.Default())
		raw := []byte("GET / \r\n")
		done, _, err := parser.Parse(raw)
		require.EqualError(t, err, status.ErrHTTPVersionNotSupported.Error())
		require.True(t, done)
	})

	t.Run("too long header key", func(t *testing.T) {
		parser, _ := getParser(config.Default())
		s := config.Default().Headers
		raw := fmt.Sprintf(
			"GET / HTTP/1.1\r\n%s: some value\r\n\r\n",
			strings.Repeat("a", s.MaxKeyLength*s.Number.Maximal+1),
		)
		_, _, err := parser.Parse([]byte(raw))
		require.EqualError(t, err, status.ErrHeaderFieldsTooLarge.Error())
	})

	t.Run("too long header value", func(t *testing.T) {
		parser, _ := getParser(config.Default())
		raw := fmt.Sprintf(
			"GET / HTTP/1.1\r\nSome-Header: %s\r\n\r\n",
			strings.Repeat("a", config.Default().Headers.MaxValueLength+1),
		)
		_, _, err := parser.Parse([]byte(raw))
		require.EqualError(t, err, status.ErrHeaderFieldsTooLarge.Error())
	})

	t.Run("too many headers", func(t *testing.T) {
		parser, _ := getParser(config.Default())
		hdrs := genHeaders(config.Default().Headers.Number.Maximal + 1)
		raw := fmt.Sprintf(
			"GET / HTTP/1.1\r\n%s\r\n\r\n",
			strings.Join(hdrs, "\r\n"),
		)
		_, _, err := parser.Parse([]byte(raw))
		require.EqualError(t, err, status.ErrTooManyHeaders.Error())
	})

	t.Run("too many Transfer-Encoding tokens", func(t *testing.T) {
		cfg := config.Default()
		cfg.Headers.MaxEncodingTokens = 3
		parser, _ := getParser(cfg)
		raw := "GET / HTTP/1.1\r\nTransfer-Encoding: gzip, deflate, br, chunked\r\n\r\n"
		done, extra, err := parser.Parse([]byte(raw))
		require.True(t, done)
		require.Empty(t, string(extra))
		require.EqualError(t, err, status.ErrTooManyEncodingTokens.Error())
	})
}

func TestParseEncoding(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		toks, err := parseEncodingString(make([]string, 0, 10), "")
		require.NoError(t, err)
		require.Empty(t, toks)
	})

	t.Run("chunked", func(t *testing.T) {
		toks, err := parseEncodingString(make([]string, 0, 10), "chunked")
		require.NoError(t, err)
		require.Equal(t, []string{"chunked"}, toks)
	})

	t.Run("gzip", func(t *testing.T) {
		toks, err := parseEncodingString(make([]string, 0, 10), "gzip")
		require.NoError(t, err)
		require.Equal(t, []string{"gzip"}, toks)
	})

	t.Run("multiple tokens", func(t *testing.T) {
		for i, tc := range []struct {
			Sample string
			Want   []string
		}{
			{"chunked,gzip", []string{"chunked", "gzip"}},
			{"gzip,chunked", []string{"gzip", "chunked"}},
			{" gzip,    chunked  ", []string{"gzip", "chunked"}},
		} {
			toks, err := parseEncodingString(make([]string, 0, 2), tc.Sample)
			if assert.NoError(t, err) {
				assert.Equal(t, tc.Want, toks, i+1)
			}
		}

		toks, err := parseEncodingString(make([]string, 0, 10), "chunked,gzip")
		require.NoError(t, err)
		require.Equal(t, []string{"chunked", "gzip"}, toks)
		toks, err = parseEncodingString(make([]string, 0, 10), "gzip,chunked")
		require.NoError(t, err)
		require.Equal(t, []string{"gzip", "chunked"}, toks)
	})

	t.Run("extra commas", func(t *testing.T) {
		for i, tc := range []string{
			", chunked",
			"chunked, ",
			"gzip,,chunked",
			"gzip, , chunked",
		} {
			_, err := parseEncodingString(make([]string, 0, 2), tc)
			assert.EqualError(t, err, status.ErrUnsupportedEncoding.Error(), i+1)
		}
	})

	t.Run("overflow tokens limit", func(t *testing.T) {
		toks, err := parseEncodingString(make([]string, 0, 1), "gzip,flate,chunked")
		require.EqualError(t, err, status.ErrTooManyEncodingTokens.Error())
		require.Nil(t, toks)
	})
}

func genHeaders(n int) (out []string) {
	for i := 0; i < n; i++ {
		out = append(out, genHeader())
	}

	return out
}

func genHeader() string {
	return fmt.Sprintf("%[1]s: %[1]s", uniuri.NewLen(16))
}
