package http1

import (
	"indigo/errors"
	"indigo/http/headers"
	methods "indigo/http/method"
	httpparser "indigo/http/parser"
	"indigo/http/proto"
	"indigo/http/url"
	settings2 "indigo/settings"
	"indigo/types"
	"testing"

	"github.com/stretchr/testify/require"
)

var (
	simpleGET = []byte("GET / HTTP/1.1\r\n\r\n")
	biggerGET = []byte("GET / HTTP/1.1\r\nHello: World!\r\n\r\n")

	biggerGETOnlyLF     = []byte("GET / HTTP/1.1\nHello: World!\n\n")
	biggerGETURLEncoded = []byte("GET /hello%20world HTTP/1.1\r\n\r\n")

	somePOST = []byte("POST / HTTP/1.1\r\nHello: World!\r\nContent-Length: 13\r\n\r\nHello, World!")
)

func getParser() (httpparser.HTTPRequestsParser, *types.Request) {
	settings := settings2.Default()
	reqHeaders := make(headers.Headers)
	manager := headers.NewManager(settings.Headers)
	request, gateway := types.NewRequest(reqHeaders, manager)
	return NewHTTPRequestsParser(
		request, gateway, nil, nil, settings,
	), request
}

func readBody(request *types.Request, ch chan []byte) {
	body, _ := request.Body()
	ch <- body
}

type testHeaders map[string]string

type wantedRequest struct {
	Method   methods.Method
	Path     url.Path
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
	rawRequest []byte, n int) {
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
			Path:     url.Path("/"),
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
			Path:     url.Path("/"),
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
			Path:     url.Path("/"),
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
			Path:     url.Path("/hello world"),
			Protocol: proto.HTTP11,
			Headers: testHeaders{
				"hello": "World!",
			},
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
				Path:     url.Path("/"),
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
				Path:     url.Path("/"),
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

		require.Equal(t, httpparser.Error, state)
		require.EqualError(t, errors.ErrBadRequest, err.Error())
	})

	t.Run("ShortInvalidMethod", func(t *testing.T) {
		parser, request := getParser()

		raw := []byte("GE / HTTP/1.1\r\n\r\n")
		ch := make(chan []byte)
		go readBody(request, ch)
		state, _, err := parser.Parse(raw)

		require.Equal(t, httpparser.Error, state)
		require.EqualError(t, errors.ErrBadRequest, err.Error())
	})

	t.Run("LongInvalidMethod", func(t *testing.T) {
		parser, request := getParser()

		raw := []byte("PATCHPOSTPUT / HTTP/1.1\r\n\r\n")
		ch := make(chan []byte)
		go readBody(request, ch)
		state, _, err := parser.Parse(raw)

		require.Equal(t, httpparser.Error, state)
		require.EqualError(t, errors.ErrBadRequest, err.Error())
	})

	t.Run("ShortInvalidProtocol", func(t *testing.T) {
		parser, request := getParser()

		raw := []byte("GET / HTT\r\n\r\n")
		ch := make(chan []byte)
		go readBody(request, ch)
		state, _, err := parser.Parse(raw)

		require.Equal(t, httpparser.Error, state)
		require.EqualError(t, err, errors.ErrUnsupportedProtocol.Error())
	})

	t.Run("LongInvalidProtocol", func(t *testing.T) {
		parser, request := getParser()

		raw := []byte("GET / HTTPS/1.1\r\n\r\n")
		ch := make(chan []byte)
		go readBody(request, ch)
		state, _, err := parser.Parse(raw)

		require.Equal(t, httpparser.Error, state)
		require.EqualError(t, err, errors.ErrUnsupportedProtocol.Error())
	})

	t.Run("UnsupportedProtocol", func(t *testing.T) {
		parser, request := getParser()

		raw := []byte("GET / HTTP/1.2\r\n\r\n")
		ch := make(chan []byte)
		go readBody(request, ch)
		state, _, err := parser.Parse(raw)

		require.Equal(t, httpparser.Error, state)
		require.EqualError(t, err, errors.ErrUnsupportedProtocol.Error())
	})

	t.Run("LFCR_CRLF", func(t *testing.T) {
		parser, request := getParser()

		raw := []byte("GET / HTTP/1.1\n\r\r\n")
		ch := make(chan []byte)
		go readBody(request, ch)
		state, _, err := parser.Parse(raw)

		require.Equal(t, httpparser.Error, state)
		require.EqualError(t, errors.ErrBadRequest, err.Error())
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

		require.Equal(t, httpparser.RequestCompleted, state)
		require.Equal(t, []byte("\r"), extra)
		require.NoError(t, err)
	})

	t.Run("HeaderWithoutColon", func(t *testing.T) {
		parser, request := getParser()

		raw := []byte("GET / HTTP/1.1\r\nsome header some value\r\n\r\n")
		ch := make(chan []byte)
		go readBody(request, ch)
		state, _, err := parser.Parse(raw)

		require.Equal(t, httpparser.Error, state)
		require.EqualError(t, errors.ErrBadRequest, err.Error())
	})

	t.Run("HeaderWithoutColon", func(t *testing.T) {
		parser, request := getParser()

		raw := []byte("GET / HTTP/1.1\r\nsome header some value\r\n\r\n")
		ch := make(chan []byte)
		go readBody(request, ch)
		state, _, err := parser.Parse(raw)

		require.Equal(t, httpparser.Error, state)
		require.EqualError(t, errors.ErrBadRequest, err.Error())
	})
}
