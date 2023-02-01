package http1

import (
	"testing"

	"github.com/indigo-web/indigo/internal/server/tcp/dummy"

	"github.com/indigo-web/indigo/http/status"
	"github.com/indigo-web/indigo/internal/pool"

	"github.com/indigo-web/indigo/http"
	"github.com/indigo-web/indigo/http/headers"
	methods "github.com/indigo-web/indigo/http/method"
	"github.com/indigo-web/indigo/http/proto"
	"github.com/indigo-web/indigo/http/query"
	"github.com/indigo-web/indigo/internal/alloc"
	httpparser "github.com/indigo-web/indigo/internal/parser"
	settings2 "github.com/indigo-web/indigo/settings"
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
	keyAllocator := alloc.NewAllocator(
		s.Headers.MaxKeyLength*s.Headers.Number.Default,
		s.Headers.MaxKeyLength*s.Headers.Number.Maximal,
	)
	valAllocator := alloc.NewAllocator(
		s.Headers.ValueSpace.Default, s.Headers.ValueSpace.Maximal,
	)
	objPool := pool.NewObjectPool[[]string](20)
	body := NewBodyReader(dummy.NewNopClient(), s.Body)
	request := http.NewRequest(
		headers.NewHeaders(nil), query.Query{}, http.NewResponse(), dummy.NewNopConn(), body,
	)
	startLineBuff := make([]byte, s.URL.MaxLength)

	return NewHTTPRequestsParser(
		request, keyAllocator, valAllocator, objPool, startLineBuff, s.Headers,
	), request
}

type wantedRequest struct {
	Method   methods.Method
	Path     string
	Protocol proto.Proto
	Headers  headers.Headers
}

func compareRequests(t *testing.T, wanted wantedRequest, actual *http.Request) {
	require.Equal(t, wanted.Method, actual.Method)
	require.Equal(t, wanted.Path, actual.Path)
	require.Equal(t, wanted.Protocol, actual.Proto)

	for key, values := range wanted.Headers.AsMap() {
		actualValues := actual.Headers.Values(key)
		require.NotNil(t, actualValues)
		require.Equal(t, values, actualValues)
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
		require.Equal(t, httpparser.RequestCompleted, state)
		require.Empty(t, extra)

		wanted := wantedRequest{
			Method:   methods.GET,
			Path:     "/",
			Protocol: proto.HTTP11,
			Headers:  headers.Headers{},
		}

		compareRequests(t, wanted, request)
		require.NoError(t, request.Reset())
		parser.Release()
	})

	t.Run("SimpleGETLeadingCRLF", func(t *testing.T) {
		state, extra, err := parser.Parse(simpleGETLeadingCRLF)
		require.NoError(t, err)
		require.Equal(t, httpparser.RequestCompleted, state)
		require.Empty(t, extra)

		wanted := wantedRequest{
			Method:   methods.GET,
			Path:     "/",
			Protocol: proto.HTTP11,
			Headers:  headers.Headers{},
		}

		compareRequests(t, wanted, request)
		require.NoError(t, request.Reset())
		parser.Release()
	})

	t.Run("BiggerGET", func(t *testing.T) {
		state, extra, err := parser.Parse(biggerGET)
		require.NoError(t, err)
		require.Equal(t, httpparser.RequestCompleted, state)
		require.Empty(t, extra)

		wanted := wantedRequest{
			Method:   methods.GET,
			Path:     "/",
			Protocol: proto.HTTP11,
			Headers: headers.NewHeaders(map[string][]string{
				"hello": {"World!"},
			}),
		}

		compareRequests(t, wanted, request)
		require.NoError(t, request.Reset())
		parser.Release()
	})

	t.Run("MultipleHeaderValues", func(t *testing.T) {
		state, extra, err := parser.Parse(multipleHeaders)
		require.NoError(t, err)
		require.Equal(t, httpparser.RequestCompleted, state)
		require.Empty(t, extra)

		wanted := wantedRequest{
			Method:   methods.GET,
			Path:     "/",
			Protocol: proto.HTTP11,
			Headers: headers.NewHeaders(map[string][]string{
				"accept": {"one,two", "three"},
			}),
		}

		compareRequests(t, wanted, request)
		require.NoError(t, request.Reset())
		parser.Release()
	})

	t.Run("BiggerGETOnlyLF", func(t *testing.T) {
		state, extra, err := parser.Parse(biggerGETOnlyLF)
		require.NoError(t, err)
		require.Equal(t, httpparser.RequestCompleted, state)
		require.Empty(t, extra)

		wanted := wantedRequest{
			Method:   methods.GET,
			Path:     "/",
			Protocol: proto.HTTP11,
			Headers: headers.NewHeaders(map[string][]string{
				"hello": {"World!"},
			}),
		}

		compareRequests(t, wanted, request)
		require.NoError(t, request.Reset())
		parser.Release()
	})

	t.Run("BiggerGET_URLEncoded", func(t *testing.T) {
		state, extra, err := parser.Parse(biggerGETURLEncoded)
		require.NoError(t, err)
		require.Equal(t, httpparser.RequestCompleted, state)
		require.Empty(t, extra)

		wanted := wantedRequest{
			Method:   methods.GET,
			Path:     "/hello world",
			Protocol: proto.HTTP11,
			Headers:  headers.Headers{},
		}

		compareRequests(t, wanted, request)
		require.NoError(t, request.Reset())
		parser.Release()
	})

	t.Run("BiggerGET_ByDifferentPartSizes", func(t *testing.T) {
		for i := 1; i < len(biggerGET); i++ {
			state, extra, err := feedPartially(parser, biggerGET, i)
			require.NoError(t, err)
			require.Empty(t, extra)
			require.Equal(t, httpparser.RequestCompleted, state)

			wanted := wantedRequest{
				Method:   methods.GET,
				Path:     "/",
				Protocol: proto.HTTP11,
				Headers: headers.NewHeaders(map[string][]string{
					"hello": {"World!"},
				}),
			}

			compareRequests(t, wanted, request)
			require.NoError(t, request.Reset())
			parser.Release()
		}
	})

	t.Run("SimpleGETWithAbsolutePath", func(t *testing.T) {
		state, extra, err := parser.Parse(simpleGETAbsPath)
		require.NoError(t, err)
		require.Equal(t, httpparser.RequestCompleted, state)
		require.Empty(t, extra)

		wanted := wantedRequest{
			Method:   methods.GET,
			Path:     "http://www.w3.org/pub/WWW/TheProject.html",
			Protocol: proto.HTTP11,
			Headers:  headers.Headers{},
		}

		compareRequests(t, wanted, request)
		require.NoError(t, request.Reset())
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
				Method:   methods.POST,
				Path:     "/",
				Protocol: proto.HTTP11,
				Headers: headers.NewHeaders(map[string][]string{
					"hello": {"World!"},
				}),
			}

			compareRequests(t, wanted, request)
			require.NoError(t, request.Reset())
			parser.Release()
		}
	})

	t.Run("SimpleGETWithQuery", func(t *testing.T) {
		state, extra, err := parser.Parse(simpleGETQuery)
		require.NoError(t, err)
		require.Equal(t, httpparser.RequestCompleted, state)
		require.Empty(t, extra)

		wanted := wantedRequest{
			Method:   methods.GET,
			Path:     "/path",
			Protocol: proto.HTTP11,
			Headers:  headers.Headers{},
		}

		compareRequests(t, wanted, request)
		require.Equal(t, "hel lo=wor ld", string(request.Query.Raw()))
		require.NoError(t, request.Reset())
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

	t.Run("ShortInvalidProtocol", func(t *testing.T) {
		parser, _ := getParser()
		raw := []byte("GET / HTT\r\n\r\n")
		state, _, err := parser.Parse(raw)
		require.EqualError(t, err, status.ErrUnsupportedProtocol.Error())
		require.Equal(t, httpparser.Error, state)
	})

	t.Run("LongInvalidProtocol", func(t *testing.T) {
		parser, _ := getParser()
		raw := []byte("GET / HTTPS/1.1\r\n\r\n")
		state, _, err := parser.Parse(raw)
		require.EqualError(t, err, status.ErrUnsupportedProtocol.Error())
		require.Equal(t, httpparser.Error, state)
	})

	t.Run("UnsupportedProtocol", func(t *testing.T) {
		parser, _ := getParser()
		raw := []byte("GET / HTTP/1.2\r\n\r\n")
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
		require.Equal(t, httpparser.RequestCompleted, state)
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
}
