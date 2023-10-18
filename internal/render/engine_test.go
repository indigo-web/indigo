package render

import (
	"bufio"
	"bytes"
	"context"
	"github.com/indigo-web/chunkedbody"
	"github.com/indigo-web/indigo/http"
	"github.com/indigo-web/indigo/http/coding"
	"github.com/indigo-web/indigo/http/headers"
	"github.com/indigo-web/indigo/http/method"
	"github.com/indigo-web/indigo/http/proto"
	"github.com/indigo-web/indigo/http/query"
	"github.com/indigo-web/indigo/http/status"
	"github.com/indigo-web/indigo/internal/parser/http1"
	"github.com/indigo-web/indigo/internal/server/tcp/dummy"
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
		context.Background(), headers.NewHeaders(), query.Query{}, http.NewResponse(), dummy.NewNopConn(),
		http.NewBody(http1.NewBodyReader(dummy.NewNopClient(), nil, coding.NewManager(0))),
		nil, false,
	)
}

type accumulativeClient struct {
	Data []byte
}

func (a *accumulativeClient) Write(b []byte) error {
	a.Data = append(a.Data, b...)
	return nil
}

func TestEngine_Write(t *testing.T) {
	request := newRequest()
	request.Method = method.GET
	stdreq, err := stdhttp.NewRequest(stdhttp.MethodGet, "/", nil)
	require.NoError(t, err)

	t.Run("NoHeaders", func(t *testing.T) {
		renderer := getEngine(nil)
		writer := new(accumulativeClient)
		require.NoError(t, renderer.Write(proto.HTTP11, request, http.NewResponse(), writer))
		resp, err := stdhttp.ReadResponse(bufio.NewReader(bytes.NewBuffer(writer.Data)), stdreq)
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

		writer := new(accumulativeClient)
		require.NoError(t, renderer.Write(proto.HTTP11, request, response, writer))
		resp, err := stdhttp.ReadResponse(bufio.NewReader(bytes.NewBuffer(writer.Data)), stdreq)
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

		writer := new(accumulativeClient)
		require.NoError(t, renderer.Write(proto.HTTP11, request, response, writer))

		r, err := stdhttp.NewRequest(stdhttp.MethodHead, "/", nil)
		require.NoError(t, err)
		resp, err := stdhttp.ReadResponse(bufio.NewReader(bytes.NewBuffer(writer.Data)), r)
		require.NoError(t, err)
		require.Equal(t, len(body), int(resp.ContentLength))
		fullBody, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		require.Empty(t, fullBody)
	})

	t.Run("HTTP/1.0_NoKeepAlive", func(t *testing.T) {
		renderer := getEngine(nil)
		response := http.NewResponse()
		err := renderer.Write(proto.HTTP10, request, response, new(accumulativeClient))
		require.EqualError(t, err, status.ErrCloseConnection.Error())
	})

	t.Run("CustomCodeAndStatus", func(t *testing.T) {
		renderer := getEngine(nil)
		response := http.NewResponse().WithCode(600)

		writer := new(accumulativeClient)
		require.NoError(t, renderer.Write(proto.HTTP11, request, response, writer))
		require.NoError(t, err)
		resp, err := stdhttp.ReadResponse(bufio.NewReader(bytes.NewBuffer(writer.Data)), stdreq)
		require.NoError(t, err)
		require.Equal(t, 600, resp.StatusCode)
	})

	t.Run("WithAttachment_KnownSize", func(t *testing.T) {
		const body = "Hello, world!"
		reader := strings.NewReader(body)
		renderer := getEngine(nil)
		response := http.NewResponse().WithAttachment(reader, reader.Len())

		writer := new(accumulativeClient)
		require.NoError(t, renderer.Write(proto.HTTP11, request, response, writer))

		resp, err := stdhttp.ReadResponse(bufio.NewReader(bytes.NewBuffer(writer.Data)), stdreq)
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

		writer := new(accumulativeClient)
		require.NoError(t, renderer.Write(proto.HTTP11, request, response, writer))
		resp, err := stdhttp.ReadResponse(bufio.NewReader(bytes.NewBuffer(writer.Data)), stdreq)
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

		writer := new(accumulativeClient)
		require.NoError(t, renderer.Write(proto.HTTP11, request, response, writer))

		resp, err := stdhttp.ReadResponse(bufio.NewReader(bytes.NewBuffer(writer.Data)), stdreq)
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

		writer := new(accumulativeClient)
		require.NoError(t, renderer.Write(proto.HTTP11, request, http.NewResponse(), writer))

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
		reader := bufio.NewReader(bytes.NewBuffer(writer.Data))
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
	t.Run("simple", func(t *testing.T) {
		reader := bytes.NewBuffer([]byte("Hello, world!"))
		wantData := "d\r\nHello, world!\r\n0\r\n\r\n"
		renderer := getEngine(nil)
		renderer.fileBuff = make([]byte, math.MaxUint16)

		writer := new(accumulativeClient)
		err := renderer.writeChunkedBody(reader, writer)
		require.NoError(t, err)
		require.Equal(t, wantData, string(writer.Data))
	})

	t.Run("long", func(t *testing.T) {
		const buffSize = 64
		parser := chunkedbody.NewParser(chunkedbody.DefaultSettings())
		payload := strings.Repeat("abcdefgh", 10*buffSize)
		reader := bytes.NewBuffer([]byte(payload))
		renderer := getEngine(nil)
		renderer.fileBuff = make([]byte, buffSize)

		writer := new(accumulativeClient)
		require.NoError(t, renderer.writeChunkedBody(reader, writer))

		var data []byte
		for len(writer.Data) > 0 {
			chunk, extra, err := parser.Parse(writer.Data, false)
			if err != nil {
				require.EqualError(t, err, io.EOF.Error())
				break
			}

			require.NoError(t, err)
			data = append(data, chunk...)
			writer.Data = extra
		}

		require.Equal(t, payload, string(data))
	})
}
