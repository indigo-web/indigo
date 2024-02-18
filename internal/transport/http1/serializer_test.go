package http1

import (
	"bufio"
	"bytes"
	"github.com/indigo-web/chunkedbody"
	"github.com/indigo-web/indigo/config"
	"github.com/indigo-web/indigo/http"
	"github.com/indigo-web/indigo/http/headers"
	"github.com/indigo-web/indigo/http/method"
	"github.com/indigo-web/indigo/http/proto"
	"github.com/indigo-web/indigo/http/query"
	"github.com/indigo-web/indigo/http/status"
	"github.com/indigo-web/indigo/internal/server/tcp/dummy"
	"github.com/stretchr/testify/require"
	"io"
	"math"
	stdhttp "net/http"
	"strings"
	"testing"
)

func getSerializer(defaultHeaders map[string]string) *Serializer {
	return NewSerializer(make([]byte, 0, 1024), 128, defaultHeaders)
}

func newRequest() *http.Request {
	return http.NewRequest(
		headers.New(), new(query.Query), http.NewResponse(), dummy.NewNopConn(),
		NewBody(dummy.NewNopClient(), nil, config.Default().Body), nil,
	)
}

type accumulativeClient struct {
	Data []byte
}

func (a *accumulativeClient) Write(b []byte) error {
	a.Data = append(a.Data, b...)
	return nil
}

func TestSerializer_Write(t *testing.T) {
	request := newRequest()
	request.Method = method.GET
	stdreq, err := stdhttp.NewRequest(stdhttp.MethodGet, "/", nil)
	require.NoError(t, err)

	t.Run("default builder", func(t *testing.T) {
		serializer := getSerializer(nil)
		writer := new(accumulativeClient)
		require.NoError(t, serializer.Write(proto.HTTP11, request, http.NewResponse(), writer))
		resp, err := stdhttp.ReadResponse(bufio.NewReader(bytes.NewBuffer(writer.Data)), stdreq)
		require.Equal(t, 200, resp.StatusCode)
		require.Equal(t, 2, len(resp.Header))
		require.Contains(t, resp.Header, "Content-Length")
		require.Contains(t, resp.Header, "Content-Type")
		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		require.Empty(t, body)
	})

	testWithHeaders := func(t *testing.T, serializer *Serializer) {
		response := http.NewResponse().
			Header("Hello", "nether").
			Header("Something", "special", "here")

		writer := new(accumulativeClient)
		require.NoError(t, serializer.Write(proto.HTTP11, request, response, writer))
		resp, err := stdhttp.ReadResponse(bufio.NewReader(bytes.NewBuffer(writer.Data)), stdreq)
		require.Equal(t, 200, resp.StatusCode)

		require.Equal(t, 1, len(resp.Header["Hello"]), resp.Header)
		require.Equal(t, "nether", resp.Header["Hello"][0], resp.Header)
		require.Equal(t, 1, len(resp.Header["Server"]), resp.Header)
		require.Equal(t, "indigo", resp.Header["Server"][0], resp.Header)
		require.Equal(t, []string{"ipsum, something else"}, resp.Header["Lorem"], resp.Header)
		require.Equal(t, []string{"special", "here"}, resp.Header["Something"], resp.Header)

		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		require.Empty(t, body)
		_ = resp.Body.Close()
	}

	t.Run("default headers", func(t *testing.T) {
		defHeaders := map[string]string{
			"Hello":  "world",
			"Server": "indigo",
			"Lorem":  "ipsum, something else",
		}
		serializer := getSerializer(defHeaders)
		testWithHeaders(t, serializer)
		testWithHeaders(t, serializer)
	})

	t.Run("HEAD request", func(t *testing.T) {
		const body = "Hello, world!"
		serializer := getSerializer(nil)
		response := http.NewResponse().String(body)
		request := newRequest()
		request.Method = method.HEAD

		writer := new(accumulativeClient)
		require.NoError(t, serializer.Write(proto.HTTP11, request, response, writer))

		r, err := stdhttp.NewRequest(stdhttp.MethodHead, "/", nil)
		require.NoError(t, err)
		resp, err := stdhttp.ReadResponse(bufio.NewReader(bytes.NewBuffer(writer.Data)), r)
		require.NoError(t, err)
		require.Equal(t, len(body), int(resp.ContentLength))
		fullBody, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		require.Empty(t, fullBody)
	})

	t.Run("HTTP/1.0 without keep-alive", func(t *testing.T) {
		serializer := getSerializer(nil)
		response := http.NewResponse()
		err := serializer.Write(proto.HTTP10, request, response, new(accumulativeClient))
		require.EqualError(t, err, status.ErrCloseConnection.Error())
	})

	t.Run("custom code and status", func(t *testing.T) {
		serializer := getSerializer(nil)
		response := http.NewResponse().Code(600)

		writer := new(accumulativeClient)
		require.NoError(t, serializer.Write(proto.HTTP11, request, response, writer))
		require.NoError(t, err)
		resp, err := stdhttp.ReadResponse(bufio.NewReader(bytes.NewBuffer(writer.Data)), stdreq)
		require.NoError(t, err)
		require.Equal(t, 600, resp.StatusCode)
	})

	t.Run("attachment with known size", func(t *testing.T) {
		const body = "Hello, world!"
		reader := strings.NewReader(body)
		serializer := getSerializer(nil)
		response := http.NewResponse().Attachment(reader, reader.Len())

		writer := new(accumulativeClient)
		require.NoError(t, serializer.Write(proto.HTTP11, request, response, writer))

		resp, err := stdhttp.ReadResponse(bufio.NewReader(bytes.NewBuffer(writer.Data)), stdreq)
		require.NoError(t, err)
		require.Equal(t, len(body), int(resp.ContentLength))
		require.Nil(t, resp.TransferEncoding)
		fullBody, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		require.Equal(t, body, string(fullBody))
	})

	t.Run("attachment with unknown size", func(t *testing.T) {
		const body = "Hello, world!"
		reader := strings.NewReader(body)
		serializer := getSerializer(nil)
		response := http.NewResponse().Attachment(reader, 0)

		writer := new(accumulativeClient)
		require.NoError(t, serializer.Write(proto.HTTP11, request, response, writer))
		resp, err := stdhttp.ReadResponse(bufio.NewReader(bytes.NewBuffer(writer.Data)), stdreq)
		require.NoError(t, err)
		require.Equal(t, 1, len(resp.TransferEncoding))
		require.Equal(t, "chunked", resp.TransferEncoding[0])
		fullBody, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		require.Equal(t, body, string(fullBody))
	})

	t.Run("attachment in respose to a HEAD request", func(t *testing.T) {
		const body = "Hello, world!"
		reader := strings.NewReader(body)
		serializer := getSerializer(nil)
		response := http.NewResponse().Attachment(reader, reader.Len())
		request.Method = method.HEAD
		stdreq, err := stdhttp.NewRequest(stdhttp.MethodHead, "/", nil)
		require.NoError(t, err)

		writer := new(accumulativeClient)
		require.NoError(t, serializer.Write(proto.HTTP11, request, response, writer))

		resp, err := stdhttp.ReadResponse(bufio.NewReader(bytes.NewBuffer(writer.Data)), stdreq)
		require.NoError(t, err)
		require.Nil(t, resp.TransferEncoding)
		require.Equal(t, len(body), int(resp.ContentLength))
		fullBody, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		require.Empty(t, string(fullBody))
	})
}

func TestSerializer_PreWrite(t *testing.T) {
	t.Run("upgrade from HTTP/1.0 to HTTP/1.1", func(t *testing.T) {
		defHeaders := map[string]string{
			"Hello": "world",
		}
		serializer := getSerializer(defHeaders)
		request := newRequest()
		request.Proto = proto.HTTP10
		request.Upgrade = proto.HTTP11
		preResponse := http.NewResponse().
			Code(status.SwitchingProtocols).
			Header("Connection", "upgrade").
			Header("Upgrade", "HTTP/1.1")
		serializer.PreWrite(request.Proto, preResponse)

		writer := new(accumulativeClient)
		require.NoError(t, serializer.Write(proto.HTTP11, request, http.NewResponse(), writer))

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

func TestSerializer_ChunkedTransfer(t *testing.T) {
	t.Run("single chunk", func(t *testing.T) {
		reader := bytes.NewBuffer([]byte("Hello, world!"))
		wantData := "d\r\nHello, world!\r\n0\r\n\r\n"
		serializer := getSerializer(nil)
		serializer.fileBuff = make([]byte, math.MaxUint16)

		writer := new(accumulativeClient)
		err := serializer.writeChunkedBody(reader, writer)
		require.NoError(t, err)
		require.Equal(t, wantData, string(writer.Data))
	})

	t.Run("long chunk into small buffer", func(t *testing.T) {
		const buffSize = 64
		parser := chunkedbody.NewParser(chunkedbody.DefaultSettings())
		payload := strings.Repeat("abcdefgh", 10*buffSize)
		reader := bytes.NewBuffer([]byte(payload))
		serializer := getSerializer(nil)
		serializer.fileBuff = make([]byte, buffSize)

		writer := new(accumulativeClient)
		require.NoError(t, serializer.writeChunkedBody(reader, writer))

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
