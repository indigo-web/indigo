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
	stdhttp "net/http"
	"testing"
)

func getEngine(defaultHeaders map[string][]string) Engine {
	return NewEngine(make([]byte, 0, 1024), nil, defaultHeaders)
}

func newRequest() *http.Request {
	return http.NewRequest(
		headers.NewHeaders(nil), query.Query{}, http.NewResponse(), dummy.NewNopConn(),
		http1.NewBodyReader(
			dummy.NewNopClient(),
			settings.Default().Body,
		),
	)
}

func TestEngine_Write(t *testing.T) {
	request := newRequest()
	r, err := stdhttp.NewRequest(stdhttp.MethodGet, "/", nil)
	require.NoError(t, err)

	t.Run("NoHeaders", func(t *testing.T) {
		renderer := getEngine(nil)
		var data []byte
		err := renderer.Write(proto.HTTP11, request, http.NewResponse(), func(b []byte) error {
			data = b
			return nil
		})
		require.NoError(t, err)
		resp, err := stdhttp.ReadResponse(bufio.NewReader(bytes.NewBuffer(data)), r)
		require.Equal(t, 200, resp.StatusCode)
		require.Equal(t, 1, len(resp.Header))
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
		resp, err := stdhttp.ReadResponse(bufio.NewReader(bytes.NewBuffer(data)), r)
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

		require.EqualError(t, err, errConnWrite.Error())
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
		resp, err := stdhttp.ReadResponse(bufio.NewReader(bytes.NewBuffer(data)), r)
		require.Equal(t, 600, resp.StatusCode)
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
