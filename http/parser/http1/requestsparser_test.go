package http1

import (
	"github.com/fakefloordiv/indigo/errors"
	"github.com/fakefloordiv/indigo/http/headers"
	methods "github.com/fakefloordiv/indigo/http/method"
	httpparser "github.com/fakefloordiv/indigo/http/parser"
	"github.com/fakefloordiv/indigo/http/proto"
	"github.com/fakefloordiv/indigo/http/url"
	settings2 "github.com/fakefloordiv/indigo/settings"
	"github.com/fakefloordiv/indigo/types"
	"testing"

	"github.com/stretchr/testify/require"
)

var (
	simpleGET = []byte("GET / HTTP/1.1\r\n\r\n")
	biggerGET = []byte("GET / HTTP/1.1\r\nHello: World!\r\n\r\n")

	biggerGETOnlyLF     = []byte("GET / HTTP/1.1\nHello: World!\n\n")
	biggerGETURLEncoded = []byte("GET /hello%20world HTTP/1.1\r\n\r\n")

	somePOST = []byte("POST / HTTP/1.1\r\nHello: World!\r\nContent-Length: 13\r\n\r\nHello, World!")

	ordinaryChunkedBody = "d\r\nHello, world!\r\n1a\r\nBut what's wrong with you?\r\nf\r\nFinally am here\r\n0\r\n\r\n"
	ordinaryChunked     = []byte("POST / HTTP/1.1\r\nTransfer-Encoding: chunked\r\n\r\n" + ordinaryChunkedBody)
)

func getParser() (httpparser.HTTPRequestsParser, *types.Request) {
	settings := settings2.Default()
	manager := headers.NewManager(settings.Headers)
	request, gateway := types.NewRequest(&manager, url.Query{})
	return NewHTTPRequestsParser(
		request, gateway, nil, nil, settings, &manager,
	), request
}

func readBody(request *types.Request, ch chan []byte) {
	body, _ := request.Body()
	ch <- body
}

type testHeaders map[string]string

type wantedRequest struct {
	Method   methods.Method
	Path     string
	Protocol proto.Proto
	Headers  testHeaders
}

func compareRequests(t *testing.T, wanted wantedRequest, actual *types.Request) {
	require.Equal(t, wanted.Method, actual.Method)
	require.Equal(t, wanted.Path, actual.Path)
	require.Equal(t, wanted.Protocol, actual.Proto)

	for key, value := range wanted.Headers {
		actualValue, found := actual.Headers[key]
		require.True(t, found)
		require.Equal(t, value, string(actualValue))
	}
}

func copySlice(src []byte) (copied []byte) {
	return append(copied, src...)
}

func splitIntoParts(req []byte, n int) (parts [][]byte) {
	for i := 0; i < len(req); i += n {
		end := i + n
		if end > len(req) {
			end = len(req)
		}

		parts = append(parts, copySlice(req[i:end]))
	}

	return parts
}

func testPartedRequest(t *testing.T, parser httpparser.HTTPRequestsParser,
	rawRequest []byte, n int,
) {
	var finalState httpparser.RequestState

	for _, chunk := range splitIntoParts(rawRequest, n) {
		state, extra, err := parser.Parse(chunk)

		for len(extra) > 0 {
			state, extra, err = parser.Parse(extra)
		}

		finalState = state
		require.NoError(t, err)
		require.Empty(t, extra)
	}

	if finalState == httpparser.RequestCompleted {
		parser.FinalizeBody()
	} else {
		require.Equalf(t, httpparser.BodyCompleted, finalState,
			"Body part size: %d", n)
	}
}

func TestHttpRequestsParser_Parse_GET(t *testing.T) {
	parser, request := getParser()

	t.Run("SimpleGET", func(t *testing.T) {
		ch := make(chan []byte)
		go readBody(request, ch)
		state, extra, err := parser.Parse(simpleGET)
		parser.FinalizeBody()

		require.NoError(t, err)
		require.Equal(t, httpparser.RequestCompleted, state)
		require.Empty(t, extra)
		require.Empty(t, <-ch)

		wanted := wantedRequest{
			Method:   methods.GET,
			Path:     "/",
			Protocol: proto.HTTP11,
			Headers:  testHeaders{},
		}

		compareRequests(t, wanted, request)
		require.NoError(t, request.Reset())
	})

	t.Run("BiggerGET", func(t *testing.T) {
		ch := make(chan []byte)
		go readBody(request, ch)
		state, extra, err := parser.Parse(biggerGET)
		parser.FinalizeBody()

		require.NoError(t, err)
		require.Equal(t, httpparser.RequestCompleted, state)
		require.Empty(t, extra)
		require.Empty(t, <-ch)

		wanted := wantedRequest{
			Method:   methods.GET,
			Path:     "/",
			Protocol: proto.HTTP11,
			Headers: testHeaders{
				"hello": "World!",
			},
		}

		compareRequests(t, wanted, request)
		require.NoError(t, request.Reset())
	})

	t.Run("BiggerGETOnlyLF", func(t *testing.T) {
		ch := make(chan []byte)
		go readBody(request, ch)
		state, extra, err := parser.Parse(biggerGETOnlyLF)
		parser.FinalizeBody()

		require.NoError(t, err)
		require.Equal(t, httpparser.RequestCompleted, state)
		require.Empty(t, extra)
		require.Empty(t, <-ch)

		wanted := wantedRequest{
			Method:   methods.GET,
			Path:     "/",
			Protocol: proto.HTTP11,
			Headers: testHeaders{
				"hello": "World!",
			},
		}

		compareRequests(t, wanted, request)
		require.NoError(t, request.Reset())
	})

	t.Run("BiggerGETURLEncoded", func(t *testing.T) {
		ch := make(chan []byte)
		go readBody(request, ch)
		state, extra, err := parser.Parse(biggerGETURLEncoded)
		parser.FinalizeBody()

		require.NoError(t, err)
		require.Equal(t, httpparser.RequestCompleted, state)
		require.Empty(t, extra)
		require.Empty(t, <-ch)

		wanted := wantedRequest{
			Method:   methods.GET,
			Path:     "/hello world",
			Protocol: proto.HTTP11,
			Headers:  testHeaders{},
		}

		compareRequests(t, wanted, request)
		require.NoError(t, request.Reset())
	})

	t.Run("BiggerGET_ByDifferentPartSizes", func(t *testing.T) {
		for i := 1; i < len(biggerGET); i++ {
			ch := make(chan []byte)
			go readBody(request, ch)
			testPartedRequest(t, parser, biggerGET, i)
			require.Empty(t, <-ch)

			wanted := wantedRequest{
				Method:   methods.GET,
				Path:     "/",
				Protocol: proto.HTTP11,
				Headers: testHeaders{
					"hello": "World!",
				},
			}

			compareRequests(t, wanted, request)
			require.NoError(t, request.Reset())
		}
	})
}

func TestHttpRequestsParser_ParsePOST(t *testing.T) {
	parser, request := getParser()

	t.Run("SomePOST_ByDifferentPartSizes", func(t *testing.T) {
		for i := 1; i < len(somePOST); i++ {
			ch := make(chan []byte)
			go readBody(request, ch)
			testPartedRequest(t, parser, somePOST, i)
			require.Equal(t, []byte("Hello, World!"), <-ch)

			wanted := wantedRequest{
				Method:   methods.POST,
				Path:     "/",
				Protocol: proto.HTTP11,
				Headers: testHeaders{
					"hello": "World!",
				},
			}

			compareRequests(t, wanted, request)
			require.NoError(t, request.Reset())
		}
	})
}

func TestHttpRequestsParser_Parse_Negative(t *testing.T) {
	t.Run("NoMethod", func(t *testing.T) {
		parser, request := getParser()

		raw := []byte(" / HTTP/1.1\r\n\r\n")
		ch := make(chan []byte)
		go readBody(request, ch)
		state, _, err := parser.Parse(raw)

		require.EqualError(t, err, errors.ErrBadRequest.Error())
		require.Equal(t, httpparser.Error, state)
	})

	t.Run("NoPath", func(t *testing.T) {
		parser, request := getParser()

		raw := []byte("GET HTTP/1.1\r\n\r\n")
		ch := make(chan []byte)
		go readBody(request, ch)
		state, _, err := parser.Parse(raw)

		require.EqualError(t, err, errors.ErrBadRequest.Error())
		require.Equal(t, httpparser.Error, state)
	})

	t.Run("PathWhitespace", func(t *testing.T) {
		parser, request := getParser()

		raw := []byte("GET  HTTP/1.1\r\n\r\n")
		ch := make(chan []byte)
		go readBody(request, ch)
		state, _, err := parser.Parse(raw)

		require.EqualError(t, err, errors.ErrBadRequest.Error())
		require.Equal(t, httpparser.Error, state)
	})

	t.Run("ShortInvalidMethod", func(t *testing.T) {
		parser, request := getParser()

		raw := []byte("GE / HTTP/1.1\r\n\r\n")
		ch := make(chan []byte)
		go readBody(request, ch)
		state, _, err := parser.Parse(raw)

		require.EqualError(t, err, errors.ErrBadRequest.Error())
		require.Equal(t, httpparser.Error, state)
	})

	t.Run("LongInvalidMethod", func(t *testing.T) {
		parser, request := getParser()

		raw := []byte("PATCHPOSTPUT / HTTP/1.1\r\n\r\n")
		ch := make(chan []byte)
		go readBody(request, ch)
		state, _, err := parser.Parse(raw)

		require.EqualError(t, err, errors.ErrBadRequest.Error())
		require.Equal(t, httpparser.Error, state)
	})

	t.Run("ShortInvalidProtocol", func(t *testing.T) {
		parser, request := getParser()

		raw := []byte("GET / HTT\r\n\r\n")
		ch := make(chan []byte)
		go readBody(request, ch)
		state, _, err := parser.Parse(raw)

		require.EqualError(t, err, errors.ErrUnsupportedProtocol.Error())
		require.Equal(t, httpparser.Error, state)
	})

	t.Run("LongInvalidProtocol", func(t *testing.T) {
		parser, request := getParser()

		raw := []byte("GET / HTTPS/1.1\r\n\r\n")
		ch := make(chan []byte)
		go readBody(request, ch)
		state, _, err := parser.Parse(raw)

		require.EqualError(t, err, errors.ErrUnsupportedProtocol.Error())
		require.Equal(t, httpparser.Error, state)
	})

	t.Run("UnsupportedProtocol", func(t *testing.T) {
		parser, request := getParser()

		raw := []byte("GET / HTTP/1.2\r\n\r\n")
		ch := make(chan []byte)
		go readBody(request, ch)
		state, _, err := parser.Parse(raw)

		require.EqualError(t, err, errors.ErrUnsupportedProtocol.Error())
		require.Equal(t, httpparser.Error, state)
	})

	t.Run("LFCR_CRLF", func(t *testing.T) {
		parser, request := getParser()

		raw := []byte("GET / HTTP/1.1\n\r\r\n")
		ch := make(chan []byte)
		go readBody(request, ch)
		state, _, err := parser.Parse(raw)

		require.EqualError(t, err, errors.ErrBadRequest.Error())
		require.Equal(t, httpparser.Error, state)
	})

	t.Run("LFCR_LFCR", func(t *testing.T) {
		// our parser is able to parse both crlf and lf splitters
		// so in example below he sees LF CRLF CR
		// the last one CR will be returned as extra-bytes

		parser, request := getParser()

		raw := []byte("GET / HTTP/1.1\n\r\n\r")
		ch := make(chan []byte)
		go readBody(request, ch)
		state, extra, err := parser.Parse(raw)

		require.Equal(t, []byte("\r"), extra)
		require.NoError(t, err)
		require.Equal(t, httpparser.RequestCompleted, state)
	})

	t.Run("HeaderWithoutColon", func(t *testing.T) {
		parser, request := getParser()

		raw := []byte("GET / HTTP/1.1\r\nsome header some value\r\n\r\n")
		ch := make(chan []byte)
		go readBody(request, ch)
		state, _, err := parser.Parse(raw)

		require.EqualError(t, errors.ErrBadRequest, err.Error())
		require.Equal(t, httpparser.Error, state)
	})

	t.Run("HeaderWithoutColon", func(t *testing.T) {
		parser, request := getParser()

		raw := []byte("GET / HTTP/1.1\r\nsome header some value\r\n\r\n")
		ch := make(chan []byte)
		go readBody(request, ch)
		state, _, err := parser.Parse(raw)

		require.EqualError(t, errors.ErrBadRequest, err.Error())
		require.Equal(t, httpparser.Error, state)
	})
}

func TestHttpRequestsParser_Chunked(t *testing.T) {
	t.Run("OrdinaryChunkedRequest", func(t *testing.T) {
		parser, request := getParser()

		ch := make(chan []byte)
		go readBody(request, ch)
		state, extra, err := parser.Parse(ordinaryChunked)
		require.NoError(t, err)
		require.Equal(t, httpparser.HeadersCompleted, state)

		state, extra, err = parser.Parse(extra)
		require.NoError(t, err)
		require.Equal(t, httpparser.BodyCompleted, state)
		require.Empty(t, extra)
		require.Equal(t, "Hello, world!But what's wrong with you?Finally am here", string(<-ch))
	})
}
