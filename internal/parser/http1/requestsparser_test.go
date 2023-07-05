package http1

import (
	"fmt"
	"github.com/dchest/uniuri"
	"github.com/indigo-web/indigo/http/decode"
	"strings"
	"testing"

	"github.com/indigo-web/indigo/internal/server/tcp/dummy"

	"github.com/indigo-web/indigo/http/status"
	"github.com/indigo-web/utils/pool"

	"github.com/indigo-web/indigo/http"
	"github.com/indigo-web/indigo/http/headers"
	"github.com/indigo-web/indigo/http/method"
	"github.com/indigo-web/indigo/http/proto"
	"github.com/indigo-web/indigo/http/query"
	httpparser "github.com/indigo-web/indigo/internal/parser"
	settings2 "github.com/indigo-web/indigo/settings"
	"github.com/indigo-web/utils/arena"
	"github.com/stretchr/testify/require"
)

var (
	simpleGET            = []byte("GET / HTTP/1.1\r\n\r\n")
	simpleGETLeadingCRLF = []byte("\r\n\r\nGET / HTTP/1.1\r\n\r\n")
	simpleGETAbsPath     = []byte("GET http://www.w3.org/pub/WWW/TheProject.html HTTP/1.1\r\n\r\n")
	biggerGET            = []byte("GET / HTTP/1.1\r\nHello: World!\r\nEaster: Egg\r\n\r\n")

	simpleGETQuery = []byte("GET /path?hel+lo=wor+ld HTTP/1.1\r\n\r\n")

	biggerGETOnlyLF     = []byte("GET / HTTP/1.1\nHello: World!\n\n")
	biggerGETURLEncoded = []byte("GET /hello%20world HTTP/1.1\r\n\r\n")

	somePOST = []byte("POST / HTTP/1.1\r\nHello: World!\r\nContent-Length: 13\r\n\r\nHello, World!")

	multipleHeaders = []byte("GET / HTTP/1.1\r\nAccept: one,two\r\nAccept: three\r\n\r\n")
)

func getParser() (httpparser.HTTPRequestsParser, *http.Request) {
	s := settings2.Default()
	keyArena := arena.NewArena(
		s.Headers.MaxKeyLength*s.Headers.Number.Default,
		s.Headers.MaxKeyLength*s.Headers.Number.Maximal,
	)
	valArena := arena.NewArena(
		s.Headers.ValueSpace.Default, s.Headers.ValueSpace.Maximal,
	)
	objPool := pool.NewObjectPool[[]string](20)
	body := NewBodyReader(dummy.NewNopClient(), s.Body)
	request := http.NewRequest(
		headers.NewHeaders(nil), query.Query{}, http.NewResponse(), dummy.NewNopConn(),
		http.NewBody(body, decode.NewDecoder()), nil, false,
	)
	startLineBuff := make([]byte, s.URL.MaxLength)

	return NewHTTPRequestsParser(
		request, *keyArena, *valArena, *objPool, startLineBuff, s.Headers,
	), request
}

type wantedRequest struct {
	Headers  *headers.Headers
	Path     string
	Fragment string
	Method   method.Method
	Protocol proto.Proto
}

func compareRequests(t *testing.T, wanted wantedRequest, actual *http.Request) {
	require.Equal(t, wanted.Method, actual.Method)
	require.Equal(t, wanted.Path, actual.Path.String)
	require.Equal(t, wanted.Fragment, actual.Path.Fragment)
	require.Equal(t, wanted.Protocol, actual.Proto)

	hdrs := wanted.Headers.Unwrap()

	for i := 0; i < len(hdrs); i += 2 {
		actualValues := actual.Headers.Values(hdrs[i])
		require.NotNil(t, actualValues)
		require.Equal(t, wanted.Headers.Values(hdrs[i]), actualValues)
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
	parser httpparser.HTTPRequestsParser, rawRequest []byte, n int,
) (state httpparser.RequestState, extra []byte, err error) {
	parts := splitIntoParts(rawRequest, n)

	for _, chunk := range parts {
		state, extra, err = parser.Parse(chunk)
		if err != nil {
			return state, extra, err
		} else if state != httpparser.Pending {
			return state, extra, err
		}

		for len(extra) > 0 {
			state, extra, err = parser.Parse(extra)
			if state != httpparser.Pending {
				return state, extra, err
			}
		}
	}

	return state, extra, nil
}

func TestHttpRequestsParser_Parse_GET(t *testing.T) {
	parser, request := getParser()

	t.Run("SimpleGET", func(t *testing.T) {
		state, extra, err := parser.Parse(simpleGET)
		require.NoError(t, err)
		require.Equal(t, httpparser.HeadersCompleted, state)
		require.Empty(t, extra)

		wanted := wantedRequest{
			Method:   method.GET,
			Path:     "/",
			Protocol: proto.HTTP11,
			Headers:  headers.NewHeaders(nil),
		}

		compareRequests(t, wanted, request)
		require.NoError(t, request.Clear())
		parser.Release()
	})

	t.Run("SimpleGETLeadingCRLF", func(t *testing.T) {
		state, extra, err := parser.Parse(simpleGETLeadingCRLF)
		require.NoError(t, err)
		require.Equal(t, httpparser.HeadersCompleted, state)
		require.Empty(t, extra)

		wanted := wantedRequest{
			Method:   method.GET,
			Path:     "/",
			Protocol: proto.HTTP11,
			Headers:  headers.NewHeaders(nil),
		}

		compareRequests(t, wanted, request)
		require.NoError(t, request.Clear())
		parser.Release()
	})

	t.Run("BiggerGET", func(t *testing.T) {
		state, extra, err := parser.Parse(biggerGET)
		require.NoError(t, err)
		require.Equal(t, httpparser.HeadersCompleted, state)
		require.Empty(t, extra)

		wanted := wantedRequest{
			Method:   method.GET,
			Path:     "/",
			Protocol: proto.HTTP11,
			Headers: headers.NewHeaders(map[string][]string{
				"hello": {"World!"},
			}),
		}

		compareRequests(t, wanted, request)
		require.NoError(t, request.Clear())
		parser.Release()
	})

	t.Run("MultipleHeaderValues", func(t *testing.T) {
		state, extra, err := parser.Parse(multipleHeaders)
		require.NoError(t, err)
		require.Equal(t, httpparser.HeadersCompleted, state)
		require.Empty(t, extra)

		wanted := wantedRequest{
			Method:   method.GET,
			Path:     "/",
			Protocol: proto.HTTP11,
			Headers: headers.NewHeaders(map[string][]string{
				"accept": {"one,two", "three"},
			}),
		}

		compareRequests(t, wanted, request)
		require.NoError(t, request.Clear())
		parser.Release()
	})

	t.Run("BiggerGETOnlyLF", func(t *testing.T) {
		state, extra, err := parser.Parse(biggerGETOnlyLF)
		require.NoError(t, err)
		require.Equal(t, httpparser.HeadersCompleted, state)
		require.Empty(t, extra)

		wanted := wantedRequest{
			Method:   method.GET,
			Path:     "/",
			Protocol: proto.HTTP11,
			Headers: headers.NewHeaders(map[string][]string{
				"hello": {"World!"},
			}),
		}

		compareRequests(t, wanted, request)
		require.NoError(t, request.Clear())
		parser.Release()
	})

	t.Run("BiggerGET_URLEncoded", func(t *testing.T) {
		state, extra, err := parser.Parse(biggerGETURLEncoded)
		require.NoError(t, err)
		require.Equal(t, httpparser.HeadersCompleted, state)
		require.Empty(t, extra)

		wanted := wantedRequest{
			Method:   method.GET,
			Path:     "/hello world",
			Protocol: proto.HTTP11,
			Headers:  headers.NewHeaders(nil),
		}

		compareRequests(t, wanted, request)
		require.NoError(t, request.Clear())
		parser.Release()
	})

	t.Run("BiggerGET_ByDifferentPartSizes", func(t *testing.T) {
		for i := 1; i < len(biggerGET); i++ {
			state, extra, err := feedPartially(parser, biggerGET, i)
			require.NoError(t, err)
			require.Empty(t, extra)
			require.Equal(t, httpparser.HeadersCompleted, state)

			wanted := wantedRequest{
				Method:   method.GET,
				Path:     "/",
				Protocol: proto.HTTP11,
				Headers: headers.NewHeaders(map[string][]string{
					"hello": {"World!"},
				}),
			}

			compareRequests(t, wanted, request)
			require.NoError(t, request.Clear())
			parser.Release()
		}
	})

	t.Run("SimpleGETWithAbsolutePath", func(t *testing.T) {
		state, extra, err := parser.Parse(simpleGETAbsPath)
		require.NoError(t, err)
		require.Equal(t, httpparser.HeadersCompleted, state)
		require.Empty(t, extra)

		wanted := wantedRequest{
			Method:   method.GET,
			Path:     "http://www.w3.org/pub/WWW/TheProject.html",
			Protocol: proto.HTTP11,
			Headers:  headers.NewHeaders(nil),
		}

		compareRequests(t, wanted, request)
		require.NoError(t, request.Clear())
		parser.Release()
	})

	t.Run("FragmentAfterPath", func(t *testing.T) {
		req := "GET /#Some%20where HTTP/1.1\r\n\r\n"
		state, extra, err := parser.Parse([]byte(req))
		require.NoError(t, err)
		require.Equal(t, httpparser.HeadersCompleted, state)
		require.Empty(t, extra)

		wanted := wantedRequest{
			Method:   method.GET,
			Path:     "/",
			Fragment: "Some where",
			Protocol: proto.HTTP11,
			Headers:  headers.NewHeaders(nil),
		}

		compareRequests(t, wanted, request)
		require.NoError(t, request.Clear())
		parser.Release()
	})

	t.Run("FragmentNoPath", func(t *testing.T) {
		req := "GET #Some%20where HTTP/1.1\r\n\r\n"
		state, extra, err := parser.Parse([]byte(req))
		require.NoError(t, err)
		require.Equal(t, httpparser.HeadersCompleted, state)
		require.Empty(t, extra)

		wanted := wantedRequest{
			Method:   method.GET,
			Path:     "/",
			Fragment: "Some where",
			Protocol: proto.HTTP11,
			Headers:  headers.NewHeaders(nil),
		}

		compareRequests(t, wanted, request)
		require.NoError(t, request.Clear())
		parser.Release()
	})

	t.Run("QueryDecode_UrlEncoded", func(t *testing.T) {
		req := "GET /?some%20where=wo%20rld#Fragment HTTP/1.1\r\n\r\n"
		state, extra, err := parser.Parse([]byte(req))
		require.NoError(t, err)
		require.Equal(t, httpparser.HeadersCompleted, state)
		require.Empty(t, extra)

		wanted := wantedRequest{
			Method:   method.GET,
			Path:     "/",
			Fragment: "Fragment",
			Protocol: proto.HTTP11,
			Headers:  headers.NewHeaders(nil),
		}

		compareRequests(t, wanted, request)
		require.Equal(t, "some where=wo rld", string(request.Path.Query.Raw()))
		require.NoError(t, request.Clear())
		parser.Release()
	})

	t.Run("TheOnlyContentLength", func(t *testing.T) {
		raw := "GET / HTTP/1.1\r\nContent-Length: 13\n\r\nHello, world!"
		state, extra, err := parser.Parse([]byte(raw))
		require.NoError(t, err)
		require.Equal(t, httpparser.HeadersCompleted, state)
		require.Equal(t, "Hello, world!", string(extra))
		require.Equal(t, 13, request.ContentLength)
		require.NoError(t, request.Clear())
		parser.Release()
	})

	t.Run("ContentLength", func(t *testing.T) {
		raw := "GET / HTTP/1.1\r\nContent-Length: 13\r\nHi-Hi: ha-ha\r\n\r\nHello, world!"
		state, extra, err := parser.Parse([]byte(raw))
		require.NoError(t, err)
		require.Equal(t, httpparser.HeadersCompleted, state)
		require.Equal(t, "Hello, world!", string(extra))
		require.Equal(t, 13, request.ContentLength)
		require.True(t, request.Headers.Has("hi-hi"))
		require.Equal(t, "ha-ha", request.Headers.Value("hi-hi"))
		require.NoError(t, request.Clear())
		parser.Release()
	})
}

func TestHttpRequestsParser_ParsePOST(t *testing.T) {
	parser, request := getParser()

	t.Run("SomePOST_ByDifferentPartSizes", func(t *testing.T) {
		for i := 1; i < len(somePOST); i++ {
			state, _, err := feedPartially(parser, somePOST, i)
			require.NoError(t, err)
			require.Equal(t, httpparser.HeadersCompleted, state)

			wanted := wantedRequest{
				Method:   method.POST,
				Path:     "/",
				Protocol: proto.HTTP11,
				Headers: headers.NewHeaders(map[string][]string{
					"hello": {"World!"},
				}),
			}

			compareRequests(t, wanted, request)
			require.NoError(t, request.Clear())
			parser.Release()
		}
	})

	t.Run("SimpleGETWithQuery", func(t *testing.T) {
		state, extra, err := parser.Parse(simpleGETQuery)
		require.NoError(t, err)
		require.Equal(t, httpparser.HeadersCompleted, state)
		require.Empty(t, extra)

		wanted := wantedRequest{
			Method:   method.GET,
			Path:     "/path",
			Protocol: proto.HTTP11,
			Headers:  headers.NewHeaders(nil),
		}

		compareRequests(t, wanted, request)
		require.Equal(t, "hel lo=wor ld", string(request.Path.Query.Raw()))
		require.NoError(t, request.Clear())
		parser.Release()
	})
}

func TestHttpRequestsParser_Parse_Negative(t *testing.T) {
	t.Run("NoMethod", func(t *testing.T) {
		parser, _ := getParser()
		raw := []byte(" / HTTP/1.1\r\n\r\n")
		state, _, err := parser.Parse(raw)
		require.EqualError(t, err, status.ErrBadRequest.Error())
		require.Equal(t, httpparser.Error, state)
	})

	t.Run("NoPath", func(t *testing.T) {
		parser, _ := getParser()
		raw := []byte("GET HTTP/1.1\r\n\r\n")
		state, _, err := parser.Parse(raw)
		require.EqualError(t, err, status.ErrBadRequest.Error())
		require.Equal(t, httpparser.Error, state)
	})

	t.Run("PathWhitespace", func(t *testing.T) {
		parser, _ := getParser()
		raw := []byte("GET  HTTP/1.1\r\n\r\n")
		state, _, err := parser.Parse(raw)
		require.EqualError(t, err, status.ErrBadRequest.Error())
		require.Equal(t, httpparser.Error, state)
	})

	t.Run("ShortInvalidMethod", func(t *testing.T) {
		parser, _ := getParser()
		raw := []byte("GE / HTTP/1.1\r\n\r\n")
		state, _, err := parser.Parse(raw)
		require.EqualError(t, err, status.ErrMethodNotImplemented.Error())
		require.Equal(t, httpparser.Error, state)
	})

	t.Run("LongInvalidMethod", func(t *testing.T) {
		parser, _ := getParser()
		raw := []byte("PATCHPOSTPUT / HTTP/1.1\r\n\r\n")
		state, _, err := parser.Parse(raw)
		require.EqualError(t, err, status.ErrBadRequest.Error())
		require.Equal(t, httpparser.Error, state)
	})

	t.Run("ShortInvalidProtocolString", func(t *testing.T) {
		parser, _ := getParser()
		raw := []byte("GET / HTT\r\n\r\n")
		state, _, err := parser.Parse(raw)
		require.EqualError(t, err, status.ErrUnsupportedProtocol.Error())
		require.Equal(t, httpparser.Error, state)
	})

	t.Run("LongInvalidProtocolString", func(t *testing.T) {
		parser, _ := getParser()
		raw := []byte("GET / HTTPS/1.1\r\n\r\n")
		state, _, err := parser.Parse(raw)
		require.EqualError(t, err, status.ErrUnsupportedProtocol.Error())
		require.Equal(t, httpparser.Error, state)
	})

	t.Run("UnsupportedProtocolVersion", func(t *testing.T) {
		parser, _ := getParser()
		raw := []byte("GET / HTTP/1.2\r\n\r\n")
		state, _, err := parser.Parse(raw)
		require.EqualError(t, err, status.ErrUnsupportedProtocol.Error())
		require.Equal(t, httpparser.Error, state)
	})

	t.Run("InvalidProtocolVersion_Minor", func(t *testing.T) {
		parser, _ := getParser()
		raw := []byte("GET / HTTP/1.x\r\n\r\n")
		state, _, err := parser.Parse(raw)
		require.EqualError(t, err, status.ErrUnsupportedProtocol.Error())
		require.Equal(t, httpparser.Error, state)
	})

	t.Run("InvalidProtocolVersion_Major", func(t *testing.T) {
		parser, _ := getParser()
		raw := []byte("GET / HTTP/x.1\r\n\r\n")
		state, _, err := parser.Parse(raw)
		require.EqualError(t, err, status.ErrUnsupportedProtocol.Error())
		require.Equal(t, httpparser.Error, state)
	})

	t.Run("LFCR_CRLF", func(t *testing.T) {
		parser, _ := getParser()
		raw := []byte("GET / HTTP/1.1\n\r\r\n")
		state, _, err := parser.Parse(raw)
		require.EqualError(t, err, status.ErrBadRequest.Error())
		require.Equal(t, httpparser.Error, state)
	})

	t.Run("LFCR_LFCR", func(t *testing.T) {
		// our parser is able to parse both crlf and lf splitters
		// so in example below he sees LF CRLF CR
		// the last one CR will be returned as extra-bytes
		parser, _ := getParser()
		raw := []byte("GET / HTTP/1.1\n\r\n\r")
		state, extra, err := parser.Parse(raw)
		require.Equal(t, []byte("\r"), extra)
		require.NoError(t, err)
		require.Equal(t, httpparser.HeadersCompleted, state)
	})

	t.Run("HeaderWithoutColon", func(t *testing.T) {
		parser, _ := getParser()
		raw := []byte("GET / HTTP/1.1\r\nsome header some value\r\n\r\n")
		state, _, err := parser.Parse(raw)
		require.EqualError(t, err, status.ErrBadRequest.Error())
		require.Equal(t, httpparser.Error, state)
	})

	t.Run("HeaderWithoutColon", func(t *testing.T) {
		parser, _ := getParser()
		raw := []byte("GET / HTTP/1.1\r\nsome header some value\r\n\r\n")
		state, _, err := parser.Parse(raw)
		require.EqualError(t, err, status.ErrBadRequest.Error())
		require.Equal(t, httpparser.Error, state)
	})

	t.Run("MajorHTTPVersionOverflow", func(t *testing.T) {
		parser, _ := getParser()
		raw := []byte("GET / HTTP/335.1\r\n\r\n")
		state, _, err := parser.Parse(raw)
		require.EqualError(t, err, status.ErrUnsupportedProtocol.Error())
		require.Equal(t, httpparser.Error, state)
	})

	t.Run("MinorHTTPVersionOverflow", func(t *testing.T) {
		parser, _ := getParser()
		raw := []byte("GET / HTTP/1.335\r\n\r\n")
		state, _, err := parser.Parse(raw)
		require.EqualError(t, err, status.ErrUnsupportedProtocol.Error())
		require.Equal(t, httpparser.Error, state)
	})

	t.Run("HeadersInvalidContentLength", func(t *testing.T) {
		parser, _ := getParser()
		raw := []byte("GET / HTTP/1.1\r\nContent-Length: 1f5\r\n\r\n")
		_, _, err := parser.Parse(raw)
		require.EqualError(t, err, status.ErrBadRequest.Error())
	})

	t.Run("SimpleRequest", func(t *testing.T) {
		// Simple Requests are not supported, because our server is
		// HTTP/1.1-oriented, and in 1.1 simple request/response is
		// something like a deprecated mechanism
		parser, _ := getParser()
		raw := []byte("GET / \r\n")
		state, _, err := parser.Parse(raw)
		require.EqualError(t, err, status.ErrBadRequest.Error())
		require.Equal(t, httpparser.Error, state)
	})

	t.Run("TooLongHeaderKey", func(t *testing.T) {
		parser, _ := getParser()
		s := settings2.Default().Headers
		raw := fmt.Sprintf(
			"GET / HTTP/1.1\r\n%s: some value\r\n\r\n",
			strings.Repeat("a", s.MaxKeyLength*s.Number.Maximal+1),
		)
		_, _, err := parser.Parse([]byte(raw))
		require.EqualError(t, err, status.ErrHeaderFieldsTooLarge.Error())
	})

	t.Run("TooLongHeaderValue", func(t *testing.T) {
		parser, _ := getParser()
		raw := fmt.Sprintf(
			"GET / HTTP/1.1\r\nSome-Header: %s\r\n\r\n",
			strings.Repeat("a", settings2.Default().Headers.MaxValueLength+1),
		)
		_, _, err := parser.Parse([]byte(raw))
		require.EqualError(t, err, status.ErrHeaderFieldsTooLarge.Error())
	})

	t.Run("TooManyHeaders", func(t *testing.T) {
		parser, _ := getParser()
		hdrs := genHeaders(settings2.Default().Headers.Number.Maximal + 1)
		raw := fmt.Sprintf(
			"GET / HTTP/1.1\r\n%s\r\n\r\n",
			strings.Join(hdrs, "\r\n"),
		)
		_, _, err := parser.Parse([]byte(raw))
		require.EqualError(t, err, status.ErrTooManyHeaders.Error())
	})
}

func TestParseTransferEncoding(t *testing.T) {
	t.Run("EmptyString", func(t *testing.T) {
		te := parseTransferEncoding("")
		require.False(t, te.Chunked)
		require.Empty(t, te.Token)
	})

	t.Run("JustChunked", func(t *testing.T) {
		te := parseTransferEncoding("chunked")
		require.True(t, te.Chunked)
		require.Empty(t, te.Token)
	})

	t.Run("JustToken", func(t *testing.T) {
		te := parseTransferEncoding("gzip")
		require.False(t, te.Chunked)
		require.Equal(t, "gzip", te.Token)
	})

	t.Run("ChunkedAndToken NoSpace", func(t *testing.T) {
		te := parseTransferEncoding("chunked,gzip")
		require.True(t, te.Chunked)
		require.Equal(t, "gzip", te.Token)

		te = parseTransferEncoding("gzip,chunked")
		require.True(t, te.Chunked)
		require.Equal(t, "gzip", te.Token)
	})

	t.Run("ChunkedAndToken WithSpace", func(t *testing.T) {
		te := parseTransferEncoding("chunked,  gzip")
		require.True(t, te.Chunked)
		require.Equal(t, "gzip", te.Token)

		te = parseTransferEncoding("gzip,  chunked")
		require.True(t, te.Chunked)
		require.Equal(t, "gzip", te.Token)
	})

	t.Run("ExtraCommas", func(t *testing.T) {
		te := parseTransferEncoding(" , chunked, gzip, ")
		require.True(t, te.Chunked)
		require.Equal(t, "gzip", te.Token)

		te = parseTransferEncoding(" , chunked")
		require.True(t, te.Chunked)
	})
}

func genHeaders(n int) (out []string) {
	for i := 0; i < n; i++ {
		out = append(out, fmt.Sprintf("%s: some value", uniuri.New()))
	}

	return out
}
