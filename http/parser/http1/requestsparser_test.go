package http1

import (
	"testing"

	"github.com/fakefloordiv/indigo/http/encodings"

	"github.com/fakefloordiv/indigo/http"
	"github.com/fakefloordiv/indigo/http/headers"
	methods "github.com/fakefloordiv/indigo/http/method"
	httpparser "github.com/fakefloordiv/indigo/http/parser"
	"github.com/fakefloordiv/indigo/http/proto"
	"github.com/fakefloordiv/indigo/http/url"
	settings2 "github.com/fakefloordiv/indigo/settings"
	"github.com/fakefloordiv/indigo/types"

	"github.com/stretchr/testify/require"
)

var (
	simpleGET            = []byte("GET / HTTP/1.1\r\n\r\n")
	simpleGETLeadingCRLF = []byte("\r\n\r\nGET / HTTP/1.1\r\n\r\n")
	simpleGETAbsPath     = []byte("GET http://www.w3.org/pub/WWW/TheProject.html HTTP/1.1\r\n\r\n")
	biggerGET            = []byte("GET / HTTP/1.1\r\nHello: World!\r\nEaster: Egg\r\n\r\n")

	headerQ       = []byte("GET / HTTP/1.1\r\nHeader: world;q=0.7,value;q=0.1\r\n\r\n")
	headerOneQ    = []byte("GET / HTTP/1.1\r\nHeader: world;q=0.7,value\r\n\r\n")
	headerSecondQ = []byte("GET / HTTP/1.1\r\nHeader: world,value;q=0.1\r\n\r\n")

	headerInvalidQMajor = []byte("GET / HTTP/1.1\r\nHeader: world;q=17.1\r\n\r\n")
	headerInvalidQNoDot = []byte("GET / HTTP/1.1\r\nHeader: world;q=171\r\n\r\n")
	headerInvalidQMinor = []byte("GET / HTTP/1.1\r\nHeader: world;q=0.17\r\n\r\n")
	headerNotQ          = []byte("GET / HTTP/1.1\r\nHeader: world;charset=utf8,value\r\n\r\n")

	simpleGETQuery = []byte("GET /path?hel+lo=wor+ld HTTP/1.1\r\n\r\n")

	biggerGETOnlyLF     = []byte("GET / HTTP/1.1\nHello: World!\n\n")
	biggerGETURLEncoded = []byte("GET /hello%20world HTTP/1.1\r\n\r\n")

	somePOST = []byte("POST / HTTP/1.1\r\nHello: World!\r\nContent-Length: 13\r\n\r\nHello, World!")

	commaSplitHeader           = []byte("GET / HTTP/1.1\r\nAccept: one,two, three\r\n\r\n")
	commaInQuotedHeaderValue   = []byte("GET / HTTP/1.1\r\nAccept: one,\"two, or more\",three\r\n\r\n")
	commaSPInQuotedHeaderValue = []byte("GET / HTTP/1.1\r\nAccept: one, \"two, or more\",three\r\n\r\n")
	quoteEscapeChar            = []byte("GET / HTTP/1.1\r\nAccept: \\\"one, two,\\\"three\r\n\r\n")

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
	settings := settings2.Default()
	manager := headers.NewManager(settings.Headers)
	request, gateway := types.NewRequest(&manager, url.Query{}, nil)
	codings := encodings.NewContentEncodings()

	return NewHTTPRequestsParser(
		request, gateway, nil, nil, settings, &manager, codings,
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

	for key, values := range wanted.Headers {
		actualValues, found := actual.Headers[key]
		require.True(t, found)
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
			Headers:  headers.Headers{},
		}

		compareRequests(t, wanted, request)
		require.NoError(t, request.Reset())
	})

	t.Run("SimpleGETLeadingCRLF", func(t *testing.T) {
		ch := make(chan []byte)
		go readBody(request, ch)
		state, extra, err := parser.Parse(simpleGETLeadingCRLF)
		parser.FinalizeBody()

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
			Headers: headers.Headers{
				"hello": []headers.Header{
					{Value: "World!", Q: 10, Charset: string(DefaultCharset)},
				},
			},
		}

		compareRequests(t, wanted, request)
		require.NoError(t, request.Reset())
	})

	commaTest := func(t *testing.T, req []byte, values []string) {
		ch := make(chan []byte)
		go readBody(request, ch)
		state, extra, err := parser.Parse(req)
		parser.FinalizeBody()

		require.NoError(t, err)
		require.Equal(t, httpparser.RequestCompleted, state)
		require.Empty(t, extra)
		require.Empty(t, <-ch)

		accept := make([]headers.Header, len(values))

		for i, value := range values {
			accept[i] = headers.Header{
				Value: value, Q: 10, Charset: string(DefaultCharset),
			}
		}

		wanted := wantedRequest{
			Method:   methods.GET,
			Path:     "/",
			Protocol: proto.HTTP11,
			Headers: headers.Headers{
				"accept": accept,
			},
		}

		compareRequests(t, wanted, request)
		require.NoError(t, request.Reset())
	}

	t.Run("CommaSplitHeaderValue", func(t *testing.T) {
		commaTest(t, commaSplitHeader, []string{"one", "two", "three"})
	})

	t.Run("QuotedCommaSplitHeaderValue", func(t *testing.T) {
		commaTest(t, commaInQuotedHeaderValue, []string{"one", "\"two, or more\"", "three"})
	})

	t.Run("SPQuotedCommaSplitHeaderValue", func(t *testing.T) {
		commaTest(t, commaSPInQuotedHeaderValue, []string{"one", "\"two, or more\"", "three"})
	})

	t.Run("EscapingQuote", func(t *testing.T) {
		commaTest(t, quoteEscapeChar, []string{"\"one", "two", "\"three"})
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
			Headers: headers.Headers{
				"hello": []headers.Header{
					{Value: "World!", Q: 10, Charset: string(DefaultCharset)},
				},
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
			Headers:  headers.Headers{},
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
				Headers: headers.Headers{
					"hello": []headers.Header{
						{Value: "World!", Q: 10, Charset: string(DefaultCharset)},
					},
				},
			}

			compareRequests(t, wanted, request)
			require.NoError(t, request.Reset())
		}
	})

	t.Run("SimpleGETWithAbsolutePath", func(t *testing.T) {
		ch := make(chan []byte)
		go readBody(request, ch)
		state, extra, err := parser.Parse(simpleGETAbsPath)
		parser.FinalizeBody()

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
	})

	t.Run("HeadersQ", func(t *testing.T) {
		ch := make(chan []byte)
		go readBody(request, ch)
		_, _, err := parser.Parse(headerQ)
		parser.FinalizeBody()
		require.NoError(t, err)
		require.Empty(t, <-ch)

		wanted := wantedRequest{
			Method:   methods.GET,
			Path:     "/",
			Protocol: proto.HTTP11,
			Headers: headers.Headers{
				"header": []headers.Header{
					{Value: "world", Q: 7, Charset: string(DefaultCharset)},
					{Value: "value", Q: 1, Charset: string(DefaultCharset)},
				},
			},
		}

		compareRequests(t, wanted, request)
		require.NoError(t, request.Reset())
	})

	t.Run("HeadersOneQ", func(t *testing.T) {
		ch := make(chan []byte)
		go readBody(request, ch)
		_, _, err := parser.Parse(headerOneQ)
		parser.FinalizeBody()
		require.NoError(t, err)
		require.Empty(t, <-ch)

		wanted := wantedRequest{
			Method:   methods.GET,
			Path:     "/",
			Protocol: proto.HTTP11,
			Headers: headers.Headers{
				"header": []headers.Header{
					{Value: "world", Q: 7, Charset: string(DefaultCharset)},
					{Value: "value", Q: 10, Charset: string(DefaultCharset)},
				},
			},
		}

		compareRequests(t, wanted, request)
		require.NoError(t, request.Reset())
	})

	t.Run("HeadersSecondQ", func(t *testing.T) {
		ch := make(chan []byte)
		go readBody(request, ch)
		_, _, err := parser.Parse(headerSecondQ)
		parser.FinalizeBody()
		require.NoError(t, err)
		require.Empty(t, <-ch)

		wanted := wantedRequest{
			Method:   methods.GET,
			Path:     "/",
			Protocol: proto.HTTP11,
			Headers: headers.Headers{
				"header": []headers.Header{
					{Value: "world", Q: 10, Charset: string(DefaultCharset)},
					{Value: "value", Q: 1, Charset: string(DefaultCharset)},
				},
			},
		}

		compareRequests(t, wanted, request)
		require.NoError(t, request.Reset())
	})

	t.Run("HeadersNotQ", func(t *testing.T) {
		ch := make(chan []byte)
		go readBody(request, ch)
		_, _, err := parser.Parse(headerNotQ)
		parser.FinalizeBody()
		require.NoError(t, err)
		require.Empty(t, <-ch)

		wanted := wantedRequest{
			Method:   methods.GET,
			Path:     "/",
			Protocol: proto.HTTP11,
			Headers: headers.Headers{
				"header": []headers.Header{
					{Value: "world", Q: 10, Charset: "utf8"},
					{Value: "value", Q: 10, Charset: string(DefaultCharset)},
				},
			},
		}

		compareRequests(t, wanted, request)
		require.NoError(t, request.Reset())
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
				Headers: headers.Headers{
					"hello": []headers.Header{
						{Value: "World!", Q: 10, Charset: string(DefaultCharset)},
					},
				},
			}

			compareRequests(t, wanted, request)
			require.NoError(t, request.Reset())
		}
	})

	t.Run("SimpleGETWithQuery", func(t *testing.T) {
		ch := make(chan []byte)
		go readBody(request, ch)
		state, extra, err := parser.Parse(simpleGETQuery)
		parser.FinalizeBody()

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

	t.Run("HeadersInvalidQMajor", func(t *testing.T) {
		parser, _ := getParser()
		_, _, err := parser.Parse(headerInvalidQMajor)
		require.EqualError(t, err, http.ErrBadRequest.Error())
	})

	t.Run("HeadersInvalidQMinor", func(t *testing.T) {
		parser, _ := getParser()
		_, _, err := parser.Parse(headerInvalidQMinor)
		require.EqualError(t, err, http.ErrBadRequest.Error())
	})

	t.Run("HeadersInvalidQNoDot", func(t *testing.T) {
		parser, _ := getParser()
		_, _, err := parser.Parse(headerInvalidQNoDot)
		require.EqualError(t, err, http.ErrBadRequest.Error())
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

	t.Run("IncompleteCharset_C", func(t *testing.T) {
		parser, request := getParser()
		req := "GET / HTTP/1.1\r\nContent-Type: text/html;c\r\n\r\n"
		_, _, err := parser.Parse([]byte(req))
		require.NoError(t, err)
		require.Equal(t, "text/html;c", request.Headers["content-type"][0].Value)
	})

	t.Run("IncompleteCharset_Ch", func(t *testing.T) {
		parser, request := getParser()
		req := "GET / HTTP/1.1\r\nContent-Type: text/html;ch\r\n\r\n"
		_, _, err := parser.Parse([]byte(req))
		require.NoError(t, err)
		require.Equal(t, "text/html;ch", request.Headers["content-type"][0].Value)
	})

	t.Run("IncompleteCharset_Cha", func(t *testing.T) {
		parser, request := getParser()
		req := "GET / HTTP/1.1\r\nContent-Type: text/html;cha\r\n\r\n"
		_, _, err := parser.Parse([]byte(req))
		require.NoError(t, err)
		require.Equal(t, "text/html;cha", request.Headers["content-type"][0].Value)
	})

	t.Run("IncompleteCharset_Char", func(t *testing.T) {
		parser, request := getParser()
		req := "GET / HTTP/1.1\r\nContent-Type: text/html;char\r\n\r\n"
		_, _, err := parser.Parse([]byte(req))
		require.NoError(t, err)
		require.Equal(t, "text/html;char", request.Headers["content-type"][0].Value)
	})

	t.Run("IncompleteCharset_Chars", func(t *testing.T) {
		parser, request := getParser()
		req := "GET / HTTP/1.1\r\nContent-Type: text/html;chars\r\n\r\n"
		_, _, err := parser.Parse([]byte(req))
		require.NoError(t, err)
		require.Equal(t, "text/html;chars", request.Headers["content-type"][0].Value)
	})

	t.Run("IncompleteCharset_Charse", func(t *testing.T) {
		parser, request := getParser()
		req := "GET / HTTP/1.1\r\nContent-Type: text/html;charse\r\n\r\n"
		_, _, err := parser.Parse([]byte(req))
		require.NoError(t, err)
		require.Equal(t, "text/html;charse", request.Headers["content-type"][0].Value)
	})

	t.Run("IncompleteCharset_Charset", func(t *testing.T) {
		parser, request := getParser()
		req := "GET / HTTP/1.1\r\nContent-Type: text/html;charset\r\n\r\n"
		_, _, err := parser.Parse([]byte(req))
		require.NoError(t, err)
		require.Equal(t, "text/html;charset", request.Headers["content-type"][0].Value)
	})

	t.Run("IncompleteCharset_char=hello", func(t *testing.T) {
		parser, request := getParser()
		req := "GET / HTTP/1.1\r\nContent-Type: text/html;char=hello\r\n\r\n"
		_, _, err := parser.Parse([]byte(req))
		require.NoError(t, err)
		require.Equal(t, "text/html;char=hello", request.Headers["content-type"][0].Value)
	})

	t.Run("CharsetBuffOverflow", func(t *testing.T) {
		parser, _ := getParser()
		req := "GET / HTTP/1.1\r\nContent-Type: text/html;charset=some-very-very-long-non-existing-charset\r\n\r\n"
		_, _, err := parser.Parse([]byte(req))
		require.EqualError(t, err, http.ErrHeaderFieldsTooLarge.Error())
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
