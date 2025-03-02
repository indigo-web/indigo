package http1

import (
	"fmt"
	"github.com/dchest/uniuri"
	"github.com/indigo-web/indigo/internal/construct"
	"github.com/indigo-web/indigo/internal/requestgen"
	"github.com/indigo-web/indigo/transport/dummy"
	"strings"
	"testing"

	"github.com/indigo-web/indigo/config"
	"github.com/indigo-web/indigo/http"
	"github.com/indigo-web/indigo/http/headers"
	"github.com/indigo-web/indigo/http/method"
	"github.com/indigo-web/indigo/http/proto"
	"github.com/indigo-web/indigo/http/status"
	"github.com/stretchr/testify/require"
)

func getParser(cfg *config.Config) (*parser, *http.Request) {
	client := dummy.NewNopClient()
	body := NewBody(client, construct.Chunked(cfg.Body), cfg.Body)
	req := construct.Request(cfg, client, body)
	suit := Initialize(cfg, nil, client, req, body)

	return suit.parser, req
}

type wantedRequest struct {
	Headers  headers.Headers
	Path     string
	Method   method.Method
	Protocol proto.Proto
}

func compareRequests(t *testing.T, wanted wantedRequest, actual *http.Request) {
	require.Equal(t, wanted.Method, actual.Method)
	require.Equal(t, wanted.Path, actual.Path)
	require.Equal(t, wanted.Protocol, actual.Proto)

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

func feedPartially(
	parser *parser, rawRequest []byte, n int,
) (state requestState, extra []byte, err error) {
	parts := splitIntoParts(rawRequest, n)

	for _, chunk := range parts {
		state, extra, err = parser.Parse(chunk)
		if err != nil {
			return state, extra, err
		} else if state != Pending {
			return state, extra, err
		}

		for len(extra) > 0 {
			state, extra, err = parser.Parse(extra)
			if state != Pending {
				return state, extra, err
			}
		}
	}

	return state, extra, nil
}

func TestHttpRequestsParser_Parse_GET(t *testing.T) {
	cfg := config.Default()
	cfg.Headers.MaxEncodingTokens = 3
	parser, request := getParser(cfg)

	t.Run("simple GET", func(t *testing.T) {
		raw := "GET / HTTP/1.1\r\n\r\n"
		state, extra, err := parser.Parse([]byte(raw))
		require.NoError(t, err)
		require.Equal(t, HeadersCompleted, state)
		require.Empty(t, extra)

		wanted := wantedRequest{
			Method:   method.GET,
			Path:     "/",
			Protocol: proto.HTTP11,
			Headers:  headers.New(),
		}

		compareRequests(t, wanted, request)
		require.NoError(t, request.Reset())
	})

	t.Run("simple GET with leading CRLF", func(t *testing.T) {
		raw := "\r\n\r\nGET / HTTP/1.1\r\n\r\n"
		state, extra, err := parser.Parse([]byte(raw))
		require.Error(t, err, status.ErrBadRequest.Error())
		require.Equal(t, Error, state)
		require.Empty(t, extra)

		require.NoError(t, request.Reset())
	})

	t.Run("normal GET", func(t *testing.T) {
		raw := "GET / HTTP/1.1\r\nHello: World!\r\nEaster: Egg\r\n\r\n"
		state, extra, err := parser.Parse([]byte(raw))
		require.NoError(t, err)
		require.Equal(t, HeadersCompleted, state)
		require.Empty(t, extra)

		wanted := wantedRequest{
			Method:   method.GET,
			Path:     "/",
			Protocol: proto.HTTP11,
			Headers: headers.NewFromMap(map[string][]string{
				"hello": {"World!"},
			}),
		}

		compareRequests(t, wanted, request)
		require.NoError(t, request.Reset())
	})

	t.Run("multiple header values", func(t *testing.T) {
		raw := "GET / HTTP/1.1\r\nAccept: one,two\r\nAccept: three\r\n\r\n"
		state, extra, err := parser.Parse([]byte(raw))
		require.NoError(t, err)
		require.Equal(t, HeadersCompleted, state)
		require.Empty(t, extra)

		wanted := wantedRequest{
			Method:   method.GET,
			Path:     "/",
			Protocol: proto.HTTP11,
			Headers: headers.NewFromMap(map[string][]string{
				"accept": {"one,two", "three"},
			}),
		}

		compareRequests(t, wanted, request)
		require.NoError(t, request.Reset())
	})

	t.Run("only lf", func(t *testing.T) {
		raw := "GET / HTTP/1.1\nHello: World!\n\n"
		state, extra, err := parser.Parse([]byte(raw))
		require.NoError(t, err)
		require.Equal(t, HeadersCompleted, state)
		require.Empty(t, extra)

		wanted := wantedRequest{
			Method:   method.GET,
			Path:     "/",
			Protocol: proto.HTTP11,
			Headers: headers.NewFromMap(map[string][]string{
				"hello": {"World!"},
			}),
		}

		compareRequests(t, wanted, request)
		require.NoError(t, request.Reset())
	})

	t.Run("escaping", func(t *testing.T) {
		tcs := []struct {
			Raw      []byte
			WantPath string
		}{
			{
				[]byte("GET /hello%2C%20world HTTP/1.1\r\n\r\n"),
				"/hello, world",
			},
			{
				requestgen.Generate(strings.Repeat("%20", 500), requestgen.Headers(10)),
				"/" + strings.Repeat(" ", 500),
			},
		}

		for i, tc := range tcs {
			state, extra, err := parser.Parse(tc.Raw)
			require.NoError(t, err, i)
			require.Equal(t, HeadersCompleted, state, i)
			require.Empty(t, extra, i)

			wanted := wantedRequest{
				Method:   method.GET,
				Path:     tc.WantPath,
				Protocol: proto.HTTP11,
				Headers:  headers.New(),
			}

			compareRequests(t, wanted, request)
			require.NoError(t, request.Reset())
		}

	})

	t.Run("fuzz GET", func(t *testing.T) {
		raw := "GET / HTTP/1.1\r\nHello: World!\r\nEaster: Egg\r\n\r\n"

		for i := 1; i < len(raw); i++ {
			state, extra, err := feedPartially(parser, []byte(raw), i)
			require.NoError(t, err, i)
			require.Empty(t, extra)
			require.Equal(t, HeadersCompleted, state)

			wanted := wantedRequest{
				Method:   method.GET,
				Path:     "/",
				Protocol: proto.HTTP11,
				Headers: headers.NewFromMap(map[string][]string{
					"hello": {"World!"},
				}),
			}

			compareRequests(t, wanted, request)
			require.NoError(t, request.Reset())
		}
	})

	t.Run("absolute path", func(t *testing.T) {
		raw := "GET http://www.w3.org/pub/WWW/TheProject.html HTTP/1.1\r\n\r\n"
		state, extra, err := parser.Parse([]byte(raw))
		require.NoError(t, err)
		require.Equal(t, HeadersCompleted, state)
		require.Empty(t, extra)

		wanted := wantedRequest{
			Method:   method.GET,
			Path:     "http://www.w3.org/pub/WWW/TheProject.html",
			Protocol: proto.HTTP11,
			Headers:  headers.New(),
		}

		compareRequests(t, wanted, request)
		require.NoError(t, request.Reset())
	})

	t.Run("content-length", func(t *testing.T) {
		raw := "GET / HTTP/1.1\r\nContent-Length: 13\n\r\nHello, world!"
		state, extra, err := parser.Parse([]byte(raw))
		require.NoError(t, err)
		require.Equal(t, HeadersCompleted, state)
		require.Equal(t, "Hello, world!", string(extra))
		require.Equal(t, 13, request.ContentLength)
		require.NoError(t, request.Reset())

		raw = "GET / HTTP/1.1\r\nContent-Length: 13\r\nHi-Hi: ha-ha\r\n\r\nHello, world!"
		state, extra, err = parser.Parse([]byte(raw))
		require.NoError(t, err)
		require.Equal(t, HeadersCompleted, state)
		require.Equal(t, "Hello, world!", string(extra))
		require.Equal(t, 13, request.ContentLength)
		require.True(t, request.Headers.Has("hi-hi"))
		require.Equal(t, "ha-ha", request.Headers.Value("hi-hi"))
		require.NoError(t, request.Reset())
	})

	t.Run("connection", func(t *testing.T) {
		raw := "GET / HTTP/1.1\r\nConnection: Keep-Alive\r\n\r\n"
		state, extra, err := parser.Parse([]byte(raw))
		require.NoError(t, err)
		require.Equal(t, HeadersCompleted, state)
		require.Empty(t, string(extra))
		require.Equal(t, "Keep-Alive", request.Connection)
		require.NoError(t, request.Reset())
	})

	t.Run("Transfer-Encoding and Content-Encoding", func(t *testing.T) {
		raw := "GET / HTTP/1.1\r\nTransfer-Encoding: gzip, deflate, chunked\r\nContent-Encoding: png, jpeg, gz\r\n\r\n"
		state, extra, err := parser.Parse([]byte(raw))
		require.NoError(t, err)
		require.Equal(t, HeadersCompleted, state)
		require.Empty(t, string(extra))
		require.Equal(t, []string{"gzip", "deflate", "chunked"}, request.Encoding.Transfer)
		require.True(t, request.Encoding.Chunked)
		require.Equal(t, []string{"png", "jpeg", "gz"}, request.Encoding.Content)
		require.NoError(t, request.Reset())
	})
}

func TestHttpRequestsParser_POST(t *testing.T) {
	parser, request := getParser(config.Default())

	t.Run("fuzz POST by different chunk sizes", func(t *testing.T) {
		raw := "POST / HTTP/1.1\r\nHello: World!\r\nContent-Length: 13\r\n\r\nHello, World!"

		for i := 1; i < len(raw); i++ {
			state, _, err := feedPartially(parser, []byte(raw), i)
			require.NoError(t, err)
			require.Equal(t, HeadersCompleted, state)

			wanted := wantedRequest{
				Method:   method.POST,
				Path:     "/",
				Protocol: proto.HTTP11,
				Headers: headers.NewFromMap(map[string][]string{
					"hello": {"World!"},
				}),
			}

			compareRequests(t, wanted, request)
			require.NoError(t, request.Reset())
		}
	})

	t.Run("query in a path", func(t *testing.T) {
		raw := "GET /path?hello=world HTTP/1.1\r\n\r\n"
		state, extra, err := parser.Parse([]byte(raw))
		require.NoError(t, err)
		require.Equal(t, HeadersCompleted, state)
		require.Empty(t, extra)

		wanted := wantedRequest{
			Method:   method.GET,
			Path:     "/path",
			Protocol: proto.HTTP11,
			Headers:  headers.New(),
		}

		compareRequests(t, wanted, request)
		require.Equal(t, "hello=world", request.Query.String())
		require.NoError(t, request.Reset())
	})
}

func TestHttpRequestsParser_Parse_Negative(t *testing.T) {
	t.Run("no method", func(t *testing.T) {
		parser, _ := getParser(config.Default())
		raw := []byte(" / HTTP/1.1\r\n\r\n")
		state, _, err := parser.Parse(raw)
		require.EqualError(t, err, status.ErrBadRequest.Error())
		require.Equal(t, Error, state)
	})

	t.Run("no path", func(t *testing.T) {
		parser, _ := getParser(config.Default())
		raw := []byte("GET HTTP/1.1\r\n\r\n")
		state, _, err := parser.Parse(raw)
		require.EqualError(t, err, status.ErrBadRequest.Error())
		require.Equal(t, Error, state)
	})

	t.Run("whitespace as a path", func(t *testing.T) {
		parser, _ := getParser(config.Default())
		raw := []byte("GET  HTTP/1.1\r\n\r\n")
		state, _, err := parser.Parse(raw)
		require.EqualError(t, err, status.ErrBadRequest.Error())
		require.Equal(t, Error, state)
	})

	t.Run("short invalid method", func(t *testing.T) {
		parser, _ := getParser(config.Default())
		raw := []byte("GE / HTTP/1.1\r\n\r\n")
		state, _, err := parser.Parse(raw)
		require.EqualError(t, err, status.ErrMethodNotImplemented.Error())
		require.Equal(t, Error, state)
	})

	t.Run("long invalid method", func(t *testing.T) {
		parser, _ := getParser(config.Default())
		raw := []byte("PATCHPOSTPUT / HTTP/1.1\r\n\r\n")
		state, _, err := parser.Parse(raw)
		require.EqualError(t, err, status.ErrMethodNotImplemented.Error())
		require.Equal(t, Error, state)
	})

	t.Run("short invalid protocol", func(t *testing.T) {
		parser, _ := getParser(config.Default())
		raw := []byte("GET / HTT\r\n\r\n")
		state, _, err := parser.Parse(raw)
		require.EqualError(t, err, status.ErrUnsupportedProtocol.Error())
		require.Equal(t, Error, state)
	})

	t.Run("long invalid protocol", func(t *testing.T) {
		parser, _ := getParser(config.Default())
		raw := []byte("GET / HTTPS/1.1\r\n\r\n")
		state, _, err := parser.Parse(raw)
		require.EqualError(t, err, status.ErrUnsupportedProtocol.Error())
		require.Equal(t, Error, state)
	})

	t.Run("unsupported minor version", func(t *testing.T) {
		parser, _ := getParser(config.Default())
		raw := []byte("GET / HTTP/1.2\r\n\r\n")
		state, _, err := parser.Parse(raw)
		require.EqualError(t, err, status.ErrUnsupportedProtocol.Error())
		require.Equal(t, Error, state)
	})

	t.Run("unsupported major version", func(t *testing.T) {
		parser, _ := getParser(config.Default())
		raw := []byte("GET / HTTP/42.1\r\n\r\n")
		state, _, err := parser.Parse(raw)
		require.EqualError(t, err, status.ErrUnsupportedProtocol.Error())
		require.Equal(t, Error, state)
	})

	t.Run("invalid minor version", func(t *testing.T) {
		parser, _ := getParser(config.Default())
		raw := []byte("GET / HTTP/1.x\r\n\r\n")
		state, _, err := parser.Parse(raw)
		require.EqualError(t, err, status.ErrUnsupportedProtocol.Error())
		require.Equal(t, Error, state)
	})

	t.Run("invalid major version", func(t *testing.T) {
		parser, _ := getParser(config.Default())
		raw := []byte("GET / HTTP/x.1\r\n\r\n")
		state, _, err := parser.Parse(raw)
		require.EqualError(t, err, status.ErrUnsupportedProtocol.Error())
		require.Equal(t, Error, state)
	})

	t.Run("lfcr crlf break sequence", func(t *testing.T) {
		parser, _ := getParser(config.Default())
		raw := []byte("GET / HTTP/1.1\n\r\r\n")
		state, _, err := parser.Parse(raw)
		require.EqualError(t, err, status.ErrBadRequest.Error())
		require.Equal(t, Error, state)
	})

	t.Run("lfcr lfcr break sequence", func(t *testing.T) {
		// our parser is able to parse both crlf and lf splitters
		// so in example below he sees LF CRLF CR
		// the last one CR will be returned as extra-bytes
		parser, _ := getParser(config.Default())
		raw := []byte("GET / HTTP/1.1\n\r\n\r")
		state, extra, err := parser.Parse(raw)
		require.Equal(t, []byte("\r"), extra)
		require.NoError(t, err)
		require.Equal(t, HeadersCompleted, state)
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
		state, _, err := parser.Parse(raw)
		require.EqualError(t, err, status.ErrUnsupportedProtocol.Error())
		require.Equal(t, Error, state)
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

	t.Run("duplicate chunked TE token", func(t *testing.T) {
		parser, _ := getParser(config.Default())
		raw := "GET / HTTP/1.1\r\nTransfer-Encoding: gzip, deflate, chunked, chunked\r\n\r\n"
		state, extra, err := parser.Parse([]byte(raw))
		require.Equal(t, Error, state)
		require.Empty(t, string(extra))
		require.EqualError(t, err, status.ErrBadEncoding.Error())
	})

	t.Run("too many Transfer-Encoding tokens", func(t *testing.T) {
		cfg := config.Default()
		cfg.Headers.MaxEncodingTokens = 3
		parser, _ := getParser(cfg)
		raw := "GET / HTTP/1.1\r\nTransfer-Encoding: gzip, deflate, br, chunked\r\n\r\n"
		state, extra, err := parser.Parse([]byte(raw))
		require.Equal(t, Error, state)
		require.Empty(t, string(extra))
		require.EqualError(t, err, status.ErrTooManyEncodingTokens.Error())
	})
}

func TestParseEncoding(t *testing.T) {
	t.Run("empty string", func(t *testing.T) {
		toks, chunked, err := parseEncodingString(make([]string, 0, 10), "")
		require.NoError(t, err)
		require.Empty(t, toks)
		require.False(t, chunked)
	})

	t.Run("only chunked", func(t *testing.T) {
		toks, chunked, err := parseEncodingString(make([]string, 0, 10), "chunked")
		require.NoError(t, err)
		require.Equal(t, []string{"chunked"}, toks)
		require.True(t, chunked)
	})

	t.Run("only gzip", func(t *testing.T) {
		toks, chunked, err := parseEncodingString(make([]string, 0, 10), "gzip")
		require.NoError(t, err)
		require.Equal(t, []string{"gzip"}, toks)
		require.False(t, chunked)
	})

	t.Run("chunked,gzip without space", func(t *testing.T) {
		toks, chunked, err := parseEncodingString(make([]string, 0, 10), "chunked,gzip")
		require.NoError(t, err)
		require.Equal(t, []string{"chunked", "gzip"}, toks)
		require.True(t, chunked)
		toks, chunked, err = parseEncodingString(make([]string, 0, 10), "gzip,chunked")
		require.NoError(t, err)
		require.Equal(t, []string{"gzip", "chunked"}, toks)
		require.True(t, chunked)
	})

	t.Run("chunked,gzip with space", func(t *testing.T) {
		toks, chunked, err := parseEncodingString(make([]string, 0, 10), "chunked,  gzip")
		require.NoError(t, err)
		require.Equal(t, []string{"chunked", "gzip"}, toks)
		require.True(t, chunked)
		toks, chunked, err = parseEncodingString(make([]string, 0, 10), "gzip,  chunked")
		require.Equal(t, []string{"gzip", "chunked"}, toks)
		require.True(t, chunked)
	})

	t.Run("extra commas", func(t *testing.T) {
		toks, chunked, err := parseEncodingString(make([]string, 0, 10), " , chunked, gzip, ")
		require.NoError(t, err)
		require.Equal(t, []string{"chunked", "gzip"}, toks)
		require.True(t, chunked)
		toks, chunked, err = parseEncodingString(make([]string, 0, 10), " , chunked")
		require.Equal(t, []string{"chunked"}, toks)
		require.True(t, chunked)
	})

	t.Run("overflow tokens limit", func(t *testing.T) {
		toks, chunked, err := parseEncodingString(make([]string, 0, 1), "gzip,flate,chunked")
		require.EqualError(t, err, status.ErrTooManyEncodingTokens.Error())
		require.Nil(t, toks)
		require.False(t, chunked)
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
