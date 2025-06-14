package http1

import (
	"bufio"
	"bytes"
	"github.com/indigo-web/indigo/config"
	"github.com/indigo-web/indigo/http"
	"github.com/indigo-web/indigo/http/cookie"
	"github.com/indigo-web/indigo/http/method"
	"github.com/indigo-web/indigo/http/proto"
	"github.com/indigo-web/indigo/http/status"
	"github.com/indigo-web/indigo/internal/construct"
	"github.com/indigo-web/indigo/transport/dummy"
	"github.com/stretchr/testify/require"
	"io"
	"math"
	stdhttp "net/http"
	"strings"
	"testing"
	"time"
)

func getSerializer(defaultHeaders map[string]string, request *http.Request, writer io.Writer) *serializer {
	return newSerializer(make([]byte, 0, 1024), 128, defaultHeaders, request, writer)
}

func newRequest() *http.Request {
	return construct.Request(config.Default(), dummy.NewNopClient())
}

type NopClientWriter struct{}

func (n NopClientWriter) Write(p []byte) (int, error) {
	return len(p), nil
}

func BenchmarkSerializer(b *testing.B) {
	defaultHeadersSmall := map[string]string{
		"Server": "indigo",
	}
	defaultHeadersMedium := map[string]string{
		"Server":           "indigo",
		"Connection":       "keep-alive",
		"Accept-Encodings": "identity",
	}
	defaultHeadersBig := map[string]string{
		"Server":           "indigo",
		"Connection":       "keep-alive",
		"Accept-Encodings": "identity",
		"Easter":           "Egg",
		"Many":             "choices, variants, ways, solutions",
		"Something":        "is not happening",
		"Talking":          "allowed",
		"Lorem":            "ipsum, doremi",
	}

	response := http.NewResponse()
	request := construct.Request(config.Default(), dummy.NewNopClient())
	client := NopClientWriter{}

	b.Run("no body no def headers", func(b *testing.B) {
		buff := make([]byte, 0, 1024)
		s := newSerializer(buff, 128, nil, request, client)
		size, err := estimateResponseSize(request, response, nil)
		require.NoError(b, err)
		b.SetBytes(size)
		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			_ = s.Write(request.Protocol, response)
		}
	})

	b.Run("with 4kb body", func(b *testing.B) {
		response := http.NewResponse().String(strings.Repeat("a", 4096))
		buff := make([]byte, 0, 8192)
		serializer := newSerializer(buff, 128, nil, request, client)
		respSize, err := estimateResponseSize(request, response, nil)
		require.NoError(b, err)
		b.SetBytes(respSize)
		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			_ = serializer.Write(request.Protocol, response)
		}
	})

	b.Run("no body 1 def header", func(b *testing.B) {
		buff := make([]byte, 0, 1024)
		serializer := newSerializer(buff, 128, defaultHeadersSmall, request, client)
		respSize, err := estimateResponseSize(request, response, defaultHeadersSmall)
		require.NoError(b, err)
		b.SetBytes(respSize)
		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			_ = serializer.Write(request.Protocol, response)
		}
	})

	b.Run("no body 3 def headers", func(b *testing.B) {
		buff := make([]byte, 0, 1024)
		serializer := newSerializer(buff, 128, defaultHeadersMedium, request, client)
		respSize, err := estimateResponseSize(request, response, defaultHeadersMedium)
		require.NoError(b, err)
		b.SetBytes(respSize)
		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			_ = serializer.Write(request.Protocol, response)
		}
	})

	b.Run("no body 8 def headers", func(b *testing.B) {
		buff := make([]byte, 0, 1024)
		serializer := newSerializer(buff, 128, defaultHeadersBig, request, client)
		respSize, err := estimateResponseSize(request, response, defaultHeadersBig)
		require.NoError(b, err)
		b.SetBytes(respSize)
		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			_ = serializer.Write(request.Protocol, response)
		}
	})

	b.Run("no body 15 headers", func(b *testing.B) {
		buff := make([]byte, 0, 1024)
		request := construct.Request(config.Default(), dummy.NewNopClient())
		request.Headers = GenerateHeaders(15)
		serializer := newSerializer(buff, 128, nil, request, client)
		size, err := estimateResponseSize(request, response, nil)
		require.NoError(b, err)
		b.SetBytes(size)
		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			_ = serializer.Write(request.Protocol, response)
		}
	})

	b.Run("pre-write", func(b *testing.B) {
		preResp := http.NewResponse().Code(status.SwitchingProtocols)
		buff := make([]byte, 0, 128)
		serializer := newSerializer(buff, 128, nil, request, client)
		respSize, err := estimatePreWriteSize(request, preResp, response)
		require.NoError(b, err)
		b.SetBytes(respSize)
		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			serializer.PreWrite(request.Protocol, preResp)
			_ = serializer.Write(request.Protocol, response)
		}
	})

	// TODO: add benchmarking chunked body
}

func estimateResponseSize(req *http.Request, resp *http.Response, defHeaders map[string]string) (int64, error) {
	writer := dummy.NewSinkholeWriter()
	serializer := newSerializer(nil, 128, defHeaders, req, writer)
	err := serializer.Write(req.Protocol, resp)

	return int64(len(writer.Data)), err
}

func estimatePreWriteSize(
	req *http.Request, preWrite, resp *http.Response,
) (int64, error) {
	writer := dummy.NewSinkholeWriter()
	serializer := newSerializer(nil, 128, nil, req, writer)
	serializer.PreWrite(req.Protocol, preWrite)
	err := serializer.Write(req.Protocol, resp)

	return int64(len(writer.Data)), err
}

func TestSerializer_Write(t *testing.T) {
	request := newRequest()
	request.Method = method.GET
	stdreq, err := stdhttp.NewRequest(stdhttp.MethodGet, "/", nil)
	require.NoError(t, err)

	t.Run("default builder", func(t *testing.T) {
		writer := dummy.NewSinkholeWriter()
		serializer := getSerializer(nil, request, writer)
		require.NoError(t, serializer.Write(proto.HTTP11, http.NewResponse()))
		resp, err := stdhttp.ReadResponse(bufio.NewReader(bytes.NewBuffer(writer.Data)), stdreq)
		require.Equal(t, 200, resp.StatusCode)
		require.Equal(t, 2, len(resp.Header))
		require.Contains(t, resp.Header, "Content-Length")
		require.Contains(t, resp.Header, "Content-Type")
		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		require.Empty(t, body)
	})

	testWithHeaders := func(t *testing.T, serializer *serializer, writer *dummy.SinkholeWriter) {
		response := http.NewResponse().
			Header("Hello", "nether").
			Header("Something", "special", "here")

		require.NoError(t, serializer.Write(proto.HTTP11, response))
		resp, err := stdhttp.ReadResponse(bufio.NewReader(bytes.NewBuffer(writer.Data)), stdreq)
		require.NoError(t, err)
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
		writer := dummy.NewSinkholeWriter()
		serializer := getSerializer(defHeaders, request, writer)
		testWithHeaders(t, serializer, writer)
		testWithHeaders(t, serializer, writer)
	})

	t.Run("HEAD request", func(t *testing.T) {
		const body = "Hello, world!"
		writer := dummy.NewSinkholeWriter()
		serializer := getSerializer(nil, request, writer)
		response := http.NewResponse().String(body)
		request := newRequest()
		request.Method = method.HEAD

		require.NoError(t, serializer.Write(proto.HTTP11, response))

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
		serializer := getSerializer(nil, request, dummy.NewSinkholeWriter())
		response := http.NewResponse()
		err := serializer.Write(proto.HTTP10, response)
		require.EqualError(t, err, status.ErrCloseConnection.Error())
	})

	t.Run("custom code and status", func(t *testing.T) {
		writer := dummy.NewSinkholeWriter()
		serializer := getSerializer(nil, request, writer)
		response := http.NewResponse().Code(600)

		require.NoError(t, serializer.Write(proto.HTTP11, response))
		require.NoError(t, err)
		resp, err := stdhttp.ReadResponse(bufio.NewReader(bytes.NewBuffer(writer.Data)), stdreq)
		require.NoError(t, err)
		require.Equal(t, 600, resp.StatusCode)
	})

	t.Run("attachment with known size", func(t *testing.T) {
		const body = "Hello, world!"
		reader := strings.NewReader(body)
		writer := dummy.NewSinkholeWriter()
		serializer := getSerializer(nil, request, writer)
		response := http.NewResponse().Attachment(reader, reader.Len())

		require.NoError(t, serializer.Write(proto.HTTP11, response))
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
		writer := dummy.NewSinkholeWriter()
		serializer := getSerializer(nil, request, writer)
		response := http.NewResponse().Attachment(reader, 0)

		require.NoError(t, serializer.Write(proto.HTTP11, response))
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
		request.Method = method.HEAD
		writer := dummy.NewSinkholeWriter()
		serializer := getSerializer(nil, request, writer)
		response := http.NewResponse().Attachment(reader, reader.Len())
		stdreq, err := stdhttp.NewRequest(stdhttp.MethodHead, "/", nil)
		require.NoError(t, err)

		require.NoError(t, serializer.Write(proto.HTTP11, response))
		resp, err := stdhttp.ReadResponse(bufio.NewReader(bytes.NewBuffer(writer.Data)), stdreq)
		require.NoError(t, err)
		require.Nil(t, resp.TransferEncoding)
		require.Equal(t, len(body), int(resp.ContentLength))
		fullBody, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		require.Empty(t, string(fullBody))
	})

	t.Run("cookies", func(t *testing.T) {
		t.Run("single pair no params", func(t *testing.T) {
			writer := dummy.NewSinkholeWriter()
			serializer := getSerializer(nil, request, writer)
			response := http.NewResponse().
				Cookie(cookie.New("hello", "world"))

			require.NoError(t, serializer.Write(proto.HTTP11, response))
			stdreq, err := stdhttp.NewRequest(stdhttp.MethodHead, "/", nil)
			require.NoError(t, err)
			resp, err := stdhttp.ReadResponse(bufio.NewReader(bytes.NewBuffer(writer.Data)), stdreq)
			require.NoError(t, err)
			c := resp.Header.Get("Set-Cookie")
			require.Equal(t, "hello=world", c)
		})

		t.Run("multiple pairs with parameters", func(t *testing.T) {
			writer := dummy.NewSinkholeWriter()
			serializer := getSerializer(nil, request, writer)
			base := cookie.Build("hello", "world").
				Path("/").
				Domain("pavlo.ooo").
				Expires(time.Date(
					2010, 5, 27, 16, 10, 32, 22,
					time.FixedZone("CEST", 0),
				)).
				SameSite(cookie.SameSiteLax).
				Secure(true).
				HttpOnly(true)

			response := http.NewResponse().
				Cookie(
					base.MaxAge(3600).Cookie(),
					base.MaxAge(-1).Cookie(),
				)

			require.NoError(t, serializer.Write(proto.HTTP11, response))
			stdreq, err := stdhttp.NewRequest(stdhttp.MethodHead, "/", nil)
			require.NoError(t, err)
			resp, err := stdhttp.ReadResponse(bufio.NewReader(bytes.NewBuffer(writer.Data)), stdreq)
			require.NoError(t, err)
			cookies := resp.Header.Values("Set-Cookie")
			require.Equal(t, 2, len(cookies), "must be only 2 cookies")
			wantCookie1 := "hello=world; Path=/; Domain=pavlo.ooo; Expires=Thu, 27 May 2010 16:10:32 GMT; " +
				"MaxAge=3600; SameSite=Lax; Secure; HttpOnly"
			wantCookie2 := "hello=world; Path=/; Domain=pavlo.ooo; Expires=Thu, 27 May 2010 16:10:32 GMT; " +
				"MaxAge=0; SameSite=Lax; Secure; HttpOnly"
			require.Equal(t, wantCookie1, cookies[0])
			require.Equal(t, wantCookie2, cookies[1])
		})
	})
}

func TestSerializer_PreWrite(t *testing.T) {
	t.Run("upgrade from HTTP/1.0 to HTTP/1.1", func(t *testing.T) {
		defHeaders := map[string]string{
			"Hello": "world",
		}
		request := newRequest()
		request.Protocol = proto.HTTP10
		request.Upgrade = proto.HTTP11
		preResponse := http.NewResponse().
			Code(status.SwitchingProtocols).
			Header("Connection", "upgrade").
			Header("Upgrade", "HTTP/1.1")
		writer := dummy.NewSinkholeWriter()
		serializer := getSerializer(defHeaders, request, writer)
		serializer.PreWrite(request.Protocol, preResponse)

		require.NoError(t, serializer.Write(proto.HTTP11, http.NewResponse()))

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
		writer := dummy.NewSinkholeWriter()
		serializer := getSerializer(nil, nil, writer)
		serializer.fileBuff = make([]byte, math.MaxUint16)

		err := serializer.writeChunkedBody(reader, writer)
		require.NoError(t, err)
		require.Equal(t, wantData, string(writer.Data))
	})

	t.Run("long chunk into small buffer", func(t *testing.T) {
		const buffSize = 64
		p := newChunkedParser()
		payload := strings.Repeat("abcdefgh", 10*buffSize)
		reader := bytes.NewBuffer([]byte(payload))
		writer := dummy.NewSinkholeWriter()
		cfg := config.Default()
		req := construct.Request(cfg, dummy.NewNopClient())
		s := getSerializer(nil, req, writer)
		s.fileBuff = make([]byte, buffSize)

		require.NoError(t, s.writeChunkedBody(reader, writer))
		out, extra, err := feed(&p, writer.Data)
		require.NoError(t, err)
		require.Empty(t, extra)
		require.Equal(t, payload, string(out))
	})
}
