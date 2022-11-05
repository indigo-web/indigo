package http1

import (
	"github.com/fakefloordiv/indigo/internal/pool"
	"testing"

	"github.com/fakefloordiv/indigo/http"
	"github.com/fakefloordiv/indigo/http/encodings"
	"github.com/fakefloordiv/indigo/http/headers"
	methods "github.com/fakefloordiv/indigo/http/method"
	"github.com/fakefloordiv/indigo/http/proto"
	"github.com/fakefloordiv/indigo/http/url"
	"github.com/fakefloordiv/indigo/internal/alloc"
	httpparser "github.com/fakefloordiv/indigo/internal/parser"
	settings2 "github.com/fakefloordiv/indigo/settings"
	"github.com/fakefloordiv/indigo/types"

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

	ordinaryChunkedBody  = "d\r\nHello, world!\r\n1a\r\nBut what's wrong with you?\r\nf\r\nFinally am here\r\n0\r\n\r\n"
	traileredChunkedBody = "7\r\nMozilla\r\n9\r\nDeveloper\r\n7\r\nNetwork\r\n0\r\nExpires: date here\r\n\r\n"
	ordinaryChunked      = []byte("POST / HTTP/1.1\r\nTransfer-Encoding: chunked\r\n\r\n" + ordinaryChunkedBody)
	chunkedWithTrailers  = []byte(
		"POST / HTTP/1.1\r\n" +
			"Transfer-Encoding: chunked\r\n" +
			"Trailer: Expires, Something-Else\r\n\r\n" +
			traileredChunkedBody,
	)
)

func getParser() (httpparser.HTTPRequestsParser, *types.Request) {
	s := settings2.Default()
	keyAllocator := alloc.NewAllocator(
		int(s.Headers.KeyLength.Maximal)*int(s.Headers.Number.Default),
		int(s.Headers.KeyLength.Maximal)*int(s.Headers.Number.Maximal),
	)
	valAllocator := alloc.NewAllocator(
		int(s.Headers.ValueSpace.Default), int(s.Headers.ValueSpace.Maximal),
	)
	objPool := pool.NewObjectPool[[]string](20)
	request, gateway := types.NewRequest(headers.NewHeaders(nil), url.Query{}, nil)
	codings := encodings.NewContentDecoders()
	startLineBuff := make([]byte, s.URL.Length.Maximal)

	return NewHTTPRequestsParser(
		request, gateway, keyAllocator, valAllocator, objPool, startLineBuff, s, codings,
	), request
}

func readBody(request *types.Request, ch chan []byte) {
	body, _ := request.Body()
	ch <- body
}

type wantedRequest struct {
	Method   methods.Method
	Path     string
	Protocol proto.Proto
	Headers  headers.Headers
}

func compareRequests(t *testing.T, wanted wantedRequest, actual *types.Request) {
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

func testPartedRequest(
	t *testing.T, parser httpparser.HTTPRequestsParser, req *types.Request, rawRequest []byte, n int,
) (body []byte) {
	var finalState httpparser.RequestState
	parts := splitIntoParts(rawRequest, n)
	bodyChan := make(chan []byte)

	for i, chunk := range parts {
		state, extra, err := parser.Parse(chunk)

		switch state {
		case httpparser.RequestCompleted:
			require.Equal(t, len(parts), i+1, "finished before whole request was fed")

			return nil
		case httpparser.HeadersCompleted:
			go readBody(req, bodyChan)
		default:
			require.NotEqual(t, httpparser.Error, state)
		}

		for len(extra) > 0 {
			state, extra, err = parser.Parse(extra)
			if state == httpparser.BodyCompleted {
				require.Equal(t, len(parts), i+1, "finished before whole request was fed")

				return <-bodyChan
			}

			require.NotEqual(t, httpparser.Error, state)
		}

		finalState = state
		require.NoError(t, err)
		require.Empty(t, extra)
	}

	if finalState != httpparser.RequestCompleted {
		require.Equalf(t, httpparser.BodyCompleted, finalState, "Body part size: %d", n)
	}

	return <-bodyChan
}

func TestHttpRequestsParser_Parse_GET(t *testing.T) {
	parser, request := getParser()

	t.Run("SimpleGET", func(t *testing.T) {
		state, extra, err := parser.Parse(simpleGET)

		require.NoError(t, err)
		require.Equal(t, httpparser.RequestCompleted, state)
		require.Empty(t, extra)
		body, err := request.Body()
		require.NoError(t, err)
		require.Empty(t, body)

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
		ch := make(chan []byte)
		go readBody(request, ch)
		state, extra, err := parser.Parse(simpleGETLeadingCRLF)

		require.NoError(t, err)
		require.Equal(t, httpparser.RequestCompleted, state)
		require.Empty(t, extra)
		require.Empty(t, <-ch)

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
		ch := make(chan []byte)
		go readBody(request, ch)
		state, extra, err := parser.Parse(biggerGET)

		require.NoError(t, err)
		require.Equal(t, httpparser.RequestCompleted, state)
		require.Empty(t, extra)
		require.Empty(t, <-ch)

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
		ch := make(chan []byte)
		go readBody(request, ch)
		state, extra, err := parser.Parse(multipleHeaders)

		require.NoError(t, err)
		require.Equal(t, httpparser.RequestCompleted, state)
		require.Empty(t, extra)
		require.Empty(t, <-ch)

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
		ch := make(chan []byte)
		go readBody(request, ch)
		state, extra, err := parser.Parse(biggerGETOnlyLF)

		require.NoError(t, err)
		require.Equal(t, httpparser.RequestCompleted, state)
		require.Empty(t, extra)
		require.Empty(t, <-ch)

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
		ch := make(chan []byte)
		go readBody(request, ch)
		state, extra, err := parser.Parse(biggerGETURLEncoded)

		require.NoError(t, err)
		require.Equal(t, httpparser.RequestCompleted, state)
		require.Empty(t, extra)
		require.Empty(t, <-ch)

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
			body := testPartedRequest(t, parser, request, biggerGET, i)
			require.Empty(t, body)

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
		ch := make(chan []byte)
		go readBody(request, ch)
		state, extra, err := parser.Parse(simpleGETAbsPath)

		require.NoError(t, err)
		require.Equal(t, httpparser.RequestCompleted, state)
		require.Empty(t, extra)
		require.Empty(t, <-ch)

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
			body := testPartedRequest(t, parser, request, somePOST, i)
			require.Equal(t, "Hello, World!", string(body))

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
		ch := make(chan []byte)
		go readBody(request, ch)
		state, extra, err := parser.Parse(simpleGETQuery)

		require.NoError(t, err)
		require.Equal(t, httpparser.RequestCompleted, state)
		require.Empty(t, extra)
		require.Empty(t, <-ch)

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
		parser, request := getParser()

		raw := []byte(" / HTTP/1.1\r\n\r\n")
		ch := make(chan []byte)
		go readBody(request, ch)
		state, _, err := parser.Parse(raw)

		require.EqualError(t, err, http.ErrBadRequest.Error())
		require.Equal(t, httpparser.Error, state)
	})

	t.Run("NoPath", func(t *testing.T) {
		parser, request := getParser()

		raw := []byte("GET HTTP/1.1\r\n\r\n")
		ch := make(chan []byte)
		go readBody(request, ch)
		state, _, err := parser.Parse(raw)

		require.EqualError(t, err, http.ErrBadRequest.Error())
		require.Equal(t, httpparser.Error, state)
	})

	t.Run("PathWhitespace", func(t *testing.T) {
		parser, request := getParser()

		raw := []byte("GET  HTTP/1.1\r\n\r\n")
		ch := make(chan []byte)
		go readBody(request, ch)
		state, _, err := parser.Parse(raw)

		require.EqualError(t, err, http.ErrBadRequest.Error())
		require.Equal(t, httpparser.Error, state)
	})

	t.Run("ShortInvalidMethod", func(t *testing.T) {
		parser, request := getParser()

		raw := []byte("GE / HTTP/1.1\r\n\r\n")
		ch := make(chan []byte)
		go readBody(request, ch)
		state, _, err := parser.Parse(raw)

		require.EqualError(t, err, http.ErrMethodNotImplemented.Error())
		require.Equal(t, httpparser.Error, state)
	})

	t.Run("LongInvalidMethod", func(t *testing.T) {
		parser, request := getParser()

		raw := []byte("PATCHPOSTPUT / HTTP/1.1\r\n\r\n")
		ch := make(chan []byte)
		go readBody(request, ch)
		state, _, err := parser.Parse(raw)

		require.EqualError(t, err, http.ErrBadRequest.Error())
		require.Equal(t, httpparser.Error, state)
	})

	t.Run("ShortInvalidProtocol", func(t *testing.T) {
		parser, request := getParser()

		raw := []byte("GET / HTT\r\n\r\n")
		ch := make(chan []byte)
		go readBody(request, ch)
		state, _, err := parser.Parse(raw)

		require.EqualError(t, err, http.ErrUnsupportedProtocol.Error())
		require.Equal(t, httpparser.Error, state)
	})

	t.Run("LongInvalidProtocol", func(t *testing.T) {
		parser, request := getParser()

		raw := []byte("GET / HTTPS/1.1\r\n\r\n")
		ch := make(chan []byte)
		go readBody(request, ch)
		state, _, err := parser.Parse(raw)

		require.EqualError(t, err, http.ErrUnsupportedProtocol.Error())
		require.Equal(t, httpparser.Error, state)
	})

	t.Run("UnsupportedProtocol", func(t *testing.T) {
		parser, request := getParser()

		raw := []byte("GET / HTTP/1.2\r\n\r\n")
		ch := make(chan []byte)
		go readBody(request, ch)
		state, _, err := parser.Parse(raw)

		require.EqualError(t, err, http.ErrUnsupportedProtocol.Error())
		require.Equal(t, httpparser.Error, state)
	})

	t.Run("LFCR_CRLF", func(t *testing.T) {
		parser, request := getParser()

		raw := []byte("GET / HTTP/1.1\n\r\r\n")
		ch := make(chan []byte)
		go readBody(request, ch)
		state, _, err := parser.Parse(raw)

		require.EqualError(t, err, http.ErrBadRequest.Error())
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

		require.EqualError(t, err, http.ErrBadRequest.Error())
		require.Equal(t, httpparser.Error, state)
	})

	t.Run("HeaderWithoutColon", func(t *testing.T) {
		parser, request := getParser()

		raw := []byte("GET / HTTP/1.1\r\nsome header some value\r\n\r\n")
		ch := make(chan []byte)
		go readBody(request, ch)
		state, _, err := parser.Parse(raw)

		require.EqualError(t, err, http.ErrBadRequest.Error())
		require.Equal(t, httpparser.Error, state)
	})

	t.Run("MajorHTTPVersionOverflow", func(t *testing.T) {
		parser, request := getParser()

		raw := []byte("GET / HTTP/335.1\r\n\r\n")
		ch := make(chan []byte)
		go readBody(request, ch)
		state, _, err := parser.Parse(raw)

		require.EqualError(t, err, http.ErrUnsupportedProtocol.Error())
		require.Equal(t, httpparser.Error, state)
	})

	t.Run("MinorHTTPVersionOverflow", func(t *testing.T) {
		parser, request := getParser()

		raw := []byte("GET / HTTP/1.335\r\n\r\n")
		ch := make(chan []byte)
		go readBody(request, ch)
		state, _, err := parser.Parse(raw)

		require.EqualError(t, err, http.ErrUnsupportedProtocol.Error())
		require.Equal(t, httpparser.Error, state)
	})

	t.Run("HeadersInvalidContentLength", func(t *testing.T) {
		parser, _ := getParser()
		req := "GET / HTTP/1.1\r\nContent-Length: 1f5\r\n\r\n"
		_, _, err := parser.Parse([]byte(req))
		require.EqualError(t, err, http.ErrBadRequest.Error())
	})

	t.Run("SimpleRequest", func(t *testing.T) {
		// Simple Requests are not supported, because our server is
		// HTTP/1.1-oriented, and in 1.1 simple request/response is
		// something like a deprecated mechanism
		parser, _ := getParser()
		req := "GET / \r\n"
		_, _, err := parser.Parse([]byte(req))
		require.EqualError(t, err, http.ErrBadRequest.Error())
	})
}

func TestHttpRequestsParser_Chunked(t *testing.T) {
	t.Run("OrdinaryChunkedRequest", func(t *testing.T) {
		parser, request := getParser()

		state, extra, err := parser.Parse(ordinaryChunked)
		require.NoError(t, err)
		require.Equal(t, httpparser.HeadersCompleted, state)

		ch := make(chan []byte)
		go readBody(request, ch)

		state, extra, err = parser.Parse(extra)
		require.NoError(t, err)
		require.Equal(t, httpparser.BodyCompleted, state)
		require.Empty(t, extra)
		require.Equal(t, "Hello, world!But what's wrong with you?Finally am here", string(<-ch))
	})

	t.Run("ChunkedWithTrailers", func(t *testing.T) {
		parser, request := getParser()

		ch := make(chan []byte)
		go readBody(request, ch)
		state, extra, err := parser.Parse(chunkedWithTrailers)
		require.NoError(t, err)
		require.Equal(t, httpparser.HeadersCompleted, state)

		state, extra, err = parser.Parse(extra)
		require.NoError(t, err)
		require.Equal(t, httpparser.BodyCompleted, state)
		require.Empty(t, extra)
		require.Equal(t, "MozillaDeveloperNetwork", string(<-ch))
	})
}
