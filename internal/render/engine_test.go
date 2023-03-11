package render

import (
	"bufio"
	"bytes"
	"github.com/indigo-web/indigo/http"
	"github.com/indigo-web/indigo/http/headers"
	method "github.com/indigo-web/indigo/http/method"
	"github.com/indigo-web/indigo/http/proto"
	"github.com/indigo-web/indigo/http/query"
	"github.com/indigo-web/indigo/http/status"
	"github.com/indigo-web/indigo/internal/parser/http1"
	"github.com/indigo-web/indigo/internal/server/tcp/dummy"
	"github.com/indigo-web/indigo/settings"
	"github.com/stretchr/testify/require"
	"io"
	"math"
	stdhttp "net/http"
	"strings"
	"testing"
)

func getEngine(defaultHeaders map[string][]string) *engine {
	return newEngine(make([]byte, 0, 1024), nil, defaultHeaders)
}

func newRequest() *http.Request {
	return http.NewRequest(
		headers.NewHeaders(nil), query.Query{}, http.NewResponse(), dummy.NewNopConn(),
		http1.NewBodyReader(
			dummy.NewNopClient(),
			settings.Default().Body,
		),
		nil, false,
	)
}

func TestEngine_Write(t *testing.T) {
	request := newRequest()
	request.Method = method.GET
	stdreq, err := stdhttp.NewRequest(stdhttp.MethodGet, "/", nil)
	require.NoError(t, err)

	t.Run("NoHeaders", func(t *testing.T) {
		renderer := getEngine(nil)
		var data []byte
		err := renderer.Write(proto.HTTP11, request, http.NewResponse(), func(b []byte) error {
			data = b
			return nil
		})
		require.NoError(t, err)
		resp, err := stdhttp.ReadResponse(bufio.NewReader(bytes.NewBuffer(data)), stdreq)
		require.Equal(t, 200, resp.StatusCode)
		require.Equal(t, 2, len(resp.Header))
		require.Contains(t, resp.Header, "Content-Length")
		require.Contains(t, resp.Header, "Content-Type")
		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		require.Empty(t, body)
	})

	testWithHeaders := func(t *testing.T, renderer Engine) {
		response := http.NewResponse().
			WithHeader("Hello", "nether").
			WithHeader("Something", "special", "here")

		var data []byte
		require.NoError(t, renderer.Write(proto.HTTP11, request, response, func(b []byte) error {
			data = b
			return nil
		}))
		resp, err := stdhttp.ReadResponse(bufio.NewReader(bytes.NewBuffer(data)), stdreq)
		require.Equal(t, 200, resp.StatusCode)

		require.Equal(t, 1, len(resp.Header["Hello"]))
		require.Equal(t, "nether", resp.Header["Hello"][0])
		require.Equal(t, 1, len(resp.Header["Server"]))
		require.Equal(t, "indigo", resp.Header["Server"][0])
		require.Equal(t, []string{"ipsum", "something else"}, resp.Header["Lorem"])
		require.Equal(t, []string{"special", "here"}, resp.Header["Something"])

		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		require.Empty(t, body)
		_ = resp.Body.Close()
	}

	t.Run("WithHeaders", func(t *testing.T) {
		defHeaders := map[string][]string{
			"Hello":  {"world"},
			"Server": {"indigo"},
			"Lorem":  {"ipsum", "something else"},
		}
		renderer := getEngine(defHeaders)
		testWithHeaders(t, renderer)
	})

	t.Run("TwiceInARow", func(t *testing.T) {
		defHeaders := map[string][]string{
			"Hello":  {"world"},
			"Server": {"indigo"},
			"Lorem":  {"ipsum", "something else"},
		}
		renderer := getEngine(defHeaders)
		testWithHeaders(t, renderer)
		testWithHeaders(t, renderer)
	})

	t.Run("HeadResponse", func(t *testing.T) {
		const body = "Hello, world!"
		renderer := getEngine(nil)
		response := http.NewResponse().WithBody(body)
		request := newRequest()
		request.Method = method.HEAD

		var data []byte
		require.NoError(t, renderer.Write(proto.HTTP11, request, response, func(b []byte) error {
			data = b
			return nil
		}))

		r, err := stdhttp.NewRequest(stdhttp.MethodHead, "/", nil)
		require.NoError(t, err)
		resp, err := stdhttp.ReadResponse(bufio.NewReader(bytes.NewBuffer(data)), r)
		require.NoError(t, err)
		require.Equal(t, len(body), int(resp.ContentLength))
		fullBody, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		require.Empty(t, fullBody)
	})

	t.Run("HTTP/1.0_NoKeepAlive", func(t *testing.T) {
		renderer := getEngine(nil)
		response := http.NewResponse()

		err := renderer.Write(proto.HTTP10, request, response, func([]byte) error {
			return nil
		})

		require.EqualError(t, err, status.ErrCloseConnection.Error())
	})

	t.Run("CustomCodeAndStatus", func(t *testing.T) {
		renderer := getEngine(nil)
		response := http.NewResponse().WithCode(600)

		var data []byte
		err := renderer.Write(proto.HTTP11, request, response, func(b []byte) error {
			data = b
			return nil
		})
		require.NoError(t, err)
		resp, err := stdhttp.ReadResponse(bufio.NewReader(bytes.NewBuffer(data)), stdreq)
		require.Equal(t, 600, resp.StatusCode)
	})

	t.Run("WithAttachment_KnownSize", func(t *testing.T) {
		const body = "Hello, world!"
		reader := strings.NewReader(body)
		renderer := getEngine(nil)
		response := http.NewResponse().WithAttachment(reader, reader.Len())

		var data []byte
		require.NoError(t, renderer.Write(proto.HTTP11, request, response, func(b []byte) error {
			data = append(data, b...)
			return nil
		}))

		resp, err := stdhttp.ReadResponse(bufio.NewReader(bytes.NewBuffer(data)), stdreq)
		require.NoError(t, err)
		require.Equal(t, len(body), int(resp.ContentLength))
		require.Nil(t, resp.TransferEncoding)
		fullBody, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		require.Equal(t, body, string(fullBody))
	})

	t.Run("WithAttachment_UnknownSize", func(t *testing.T) {
		const body = "Hello, world!"
		reader := strings.NewReader(body)
		renderer := getEngine(nil)
		response := http.NewResponse().WithAttachment(reader, 0)

		var data []byte
		require.NoError(t, renderer.Write(proto.HTTP11, request, response, func(b []byte) error {
			data = append(data, b...)
			return nil
		}))

		resp, err := stdhttp.ReadResponse(bufio.NewReader(bytes.NewBuffer(data)), stdreq)
		require.NoError(t, err)
		require.Equal(t, 1, len(resp.TransferEncoding))
		require.Equal(t, "chunked", resp.TransferEncoding[0])
		fullBody, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		require.Equal(t, body, string(fullBody))
	})

	t.Run("WithAttachment_HeadRequest", func(t *testing.T) {
		const body = "Hello, world!"
		reader := strings.NewReader(body)
		renderer := getEngine(nil)
		response := http.NewResponse().WithAttachment(reader, reader.Len())
		request.Method = method.HEAD
		stdreq, err := stdhttp.NewRequest(stdhttp.MethodHead, "/", nil)
		require.NoError(t, err)

		var data []byte
		require.NoError(t, renderer.Write(proto.HTTP11, request, response, func(b []byte) error {
			data = append(data, b...)
			return nil
		}))

		resp, err := stdhttp.ReadResponse(bufio.NewReader(bytes.NewBuffer(data)), stdreq)
		require.NoError(t, err)
		require.Nil(t, resp.TransferEncoding)
		require.Equal(t, len(body), int(resp.ContentLength))
		fullBody, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		require.Empty(t, string(fullBody))
	})
}

func TestEngine_PreWrite(t *testing.T) {
	t.Run("Simple", func(t *testing.T) {
		defHeaders := map[string][]string{
			"Hello": {"world"},
		}
		renderer := getEngine(defHeaders)
		request := newRequest()
		request.Proto = proto.HTTP10
		request.Upgrade = proto.HTTP11
		preResponse := http.NewResponse().
			WithCode(status.SwitchingProtocols).
			WithHeader("Connection", "upgrade").
			WithHeader("Upgrade", "HTTP/1.1")
		renderer.PreWrite(request.Proto, preResponse)

		var data []byte
		require.NoError(t, renderer.Write(request.Upgrade, request, http.NewResponse(), func(b []byte) error {
			data = b
			return nil
		}))

		r := &stdhttp.Request{
			Method:     stdhttp.MethodGet,
			Proto:      "HTTP/1.0",
			ProtoMajor: 1,
			ProtoMinor: 0,
			Header: map[string][]string{
				"Connection": {"upgrade"},
				"Upgrade":    {"HTTP/1.1"},
			},
			RemoteAddr: "",
			RequestURI: "/",
		}
		reader := bufio.NewReader(bytes.NewBuffer(data))
		resp, err := stdhttp.ReadResponse(reader, r)
		require.NoError(t, err)
		require.Equal(t, 101, resp.StatusCode)
		require.Contains(t, resp.Header, "Hello")
		require.Equal(t, []string{"world"}, resp.Header["Hello"])
		require.Equal(t, "HTTP/1.0", resp.Proto)

		resp, err = stdhttp.ReadResponse(reader, r)
		require.NoError(t, err)
		require.Equal(t, 200, resp.StatusCode)
		require.Equal(t, "HTTP/1.1", resp.Proto)
	})
}

func TestEngine_ChunkedTransfer(t *testing.T) {
	t.Run("Simple", func(t *testing.T) {
		reader := bytes.NewBuffer([]byte("Hello, world!"))
		wantData := "d\r\nHello, world!\r\n0\r\n\r\n"
		renderer := getEngine(nil)
		renderer.fileBuff = make([]byte, math.MaxUint16)

		var data []byte
		err := renderer.writeChunkedBody(reader, func(b []byte) error {
			data = append(data, b...)
			return nil
		})
		require.NoError(t, err)
		require.Equal(t, wantData, string(data))
	})

	t.Run("Long", func(t *testing.T) {
		const buffSize = 64
		parser := http1.NewChunkedBodyParser(settings.Default().Body)
		payload := strings.Repeat("abcdefgh", 10*buffSize)
		reader := bytes.NewBuffer([]byte(payload))
		renderer := getEngine(nil)
		renderer.fileBuff = make([]byte, buffSize)

		var data []byte
		err := renderer.writeChunkedBody(reader, func(b []byte) error {
			for len(b) > 0 {
				chunk, extra, err := parser.Parse(b, false)
				if err == io.EOF {
					return nil
				}

				require.NoError(t, err)
				data = append(data, chunk...)
				b = extra
			}

			return nil
		})
		require.NoError(t, err)
		require.Equal(t, payload, string(data))
	})
}
