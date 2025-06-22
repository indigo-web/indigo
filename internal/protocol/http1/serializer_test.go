package http1

import (
	"bufio"
	"bytes"
	"github.com/indigo-web/indigo/config"
	"github.com/indigo-web/indigo/http"
	"github.com/indigo-web/indigo/http/cookie"
	"github.com/indigo-web/indigo/http/method"
	"github.com/indigo-web/indigo/http/mime"
	"github.com/indigo-web/indigo/http/proto"
	"github.com/indigo-web/indigo/http/status"
	"github.com/indigo-web/indigo/internal/construct"
	"github.com/indigo-web/indigo/transport/dummy"
	"github.com/stretchr/testify/require"
	"io"
	"net"
	stdhttp "net/http"
	"strings"
	"testing"
	"time"
)

func getSerializer(defaultHeaders map[string]string, request *http.Request, w writer) *serializer {
	return newSerializer(config.Default(), w, make([]byte, 0, 1024), defaultHeaders, request)
}

func newRequest() *http.Request {
	return construct.Request(config.Default(), dummy.NewNopClient())
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
	request := newRequest()
	w := dummy.NewNopClient()

	b.Run("no body no def headers", func(b *testing.B) {
		buff := make([]byte, 0, 1024)
		s := newSerializer(config.Default(), w, buff, nil, request)
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
		s := newSerializer(config.Default(), w, buff, nil, request)
		respSize, err := estimateResponseSize(request, response, nil)
		require.NoError(b, err)
		b.SetBytes(respSize)
		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			_ = s.Write(request.Protocol, response)
		}
	})

	b.Run("no body 1 def header", func(b *testing.B) {
		buff := make([]byte, 0, 1024)
		s := newSerializer(config.Default(), w, buff, defaultHeadersSmall, request)
		respSize, err := estimateResponseSize(request, response, defaultHeadersSmall)
		require.NoError(b, err)
		b.SetBytes(respSize)
		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			_ = s.Write(request.Protocol, response)
		}
	})

	b.Run("no body 3 def headers", func(b *testing.B) {
		buff := make([]byte, 0, 1024)
		s := newSerializer(config.Default(), w, buff, defaultHeadersMedium, request)
		respSize, err := estimateResponseSize(request, response, defaultHeadersMedium)
		require.NoError(b, err)
		b.SetBytes(respSize)
		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			_ = s.Write(request.Protocol, response)
		}
	})

	b.Run("no body 8 def headers", func(b *testing.B) {
		buff := make([]byte, 0, 1024)
		s := newSerializer(config.Default(), w, buff, defaultHeadersBig, request)
		respSize, err := estimateResponseSize(request, response, defaultHeadersBig)
		require.NoError(b, err)
		b.SetBytes(respSize)
		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			_ = s.Write(request.Protocol, response)
		}
	})

	b.Run("no body 15 headers", func(b *testing.B) {
		buff := make([]byte, 0, 1024)
		request := newRequest()
		request.Headers = GenerateHeaders(15)
		s := newSerializer(config.Default(), w, buff, nil, request)
		size, err := estimateResponseSize(request, response, nil)
		require.NoError(b, err)
		b.SetBytes(size)
		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			_ = s.Write(request.Protocol, response)
		}
	})

	b.Run("pre-write", func(b *testing.B) {
		buff := make([]byte, 0, 128)
		s := newSerializer(config.Default(), w, buff, nil, request)
		request.Upgrade = proto.HTTP11
		respSize, err := estimateUpgradeSize(request, response)
		require.NoError(b, err)
		b.SetBytes(respSize)
		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			s.Upgrade()
			_ = s.Write(request.Upgrade, response)
		}
	})

	// TODO: add benchmarking chunked body
}

func estimateResponseSize(req *http.Request, resp *http.Response, defHeaders map[string]string) (int64, error) {
	writer := new(JournalingClient)
	s := newSerializer(config.Default(), writer, make([]byte, 0, 128), defHeaders, req)
	err := s.Write(req.Protocol, resp)

	return int64(len(writer.Data)), err
}

func estimateUpgradeSize(
	req *http.Request, resp *http.Response,
) (int64, error) {
	writer := new(JournalingClient)
	s := newSerializer(config.Default(), writer, make([]byte, 0, 128), nil, req)
	s.Upgrade()
	err := s.Write(req.Protocol, resp)

	return int64(len(writer.Data)), err
}

func TestSerializer(t *testing.T) {
	request := newRequest()
	request.Method = method.GET

	t.Run("default builder", func(t *testing.T) {
		writer := new(JournalingClient)
		s := getSerializer(nil, request, writer)
		require.NoError(t, s.Write(proto.HTTP11, http.NewResponse()))

		resp, err := parseHTTP11Response("GET", writer)
		require.NoError(t, err)
		require.Equal(t, 200, resp.StatusCode)
		require.Equal(t, 2, len(resp.Header))
		require.Contains(t, resp.Header, "Content-Length")
		require.Contains(t, resp.Header, "Content-Type")
		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		require.Empty(t, body)
	})

	testWithHeaders := func(t *testing.T, s *serializer, writer *JournalingClient) {
		response := http.NewResponse().
			Header("Hello", "nether").
			Header("Something", "special", "here")

		require.NoError(t, s.Write(proto.HTTP11, response))
		resp, err := parseHTTP11Response("GET", writer)
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
		require.Empty(t, string(body))
		_ = resp.Body.Close()
	}

	t.Run("default headers", func(t *testing.T) {
		defHeaders := map[string]string{
			"Hello":  "world",
			"Server": "indigo",
			"Lorem":  "ipsum, something else",
		}
		writer := new(JournalingClient)
		s := getSerializer(defHeaders, request, writer)
		testWithHeaders(t, s, writer)
		testWithHeaders(t, s, writer)
	})

	t.Run("HTTP/1.0 without keep-alive", func(t *testing.T) {
		serializer := getSerializer(nil, request, new(JournalingClient))
		response := http.NewResponse()
		err := serializer.Write(proto.HTTP10, response)
		require.EqualError(t, err, status.ErrCloseConnection.Error())
	})

	t.Run("custom code and status", func(t *testing.T) {
		writer := new(JournalingClient)
		serializer := getSerializer(nil, request, writer)
		response := http.NewResponse().Code(600)

		require.NoError(t, serializer.Write(proto.HTTP11, response))
		resp, err := parseHTTP11Response("GET", writer)
		require.NoError(t, err)
		require.Equal(t, 600, resp.StatusCode)
	})

	t.Run("cookies", func(t *testing.T) {
		t.Run("single pair no params", func(t *testing.T) {
			writer := new(JournalingClient)
			serializer := getSerializer(nil, request, writer)
			response := http.NewResponse().
				Cookie(cookie.New("hello", "world"))

			require.NoError(t, serializer.Write(proto.HTTP11, response))
			resp, err := parseHTTP11Response("HEAD", writer)
			require.NoError(t, err)
			c := resp.Header.Get("Set-Cookie")
			require.Equal(t, "hello=world", c)
		})

		t.Run("multiple pairs with parameters", func(t *testing.T) {
			writer := new(JournalingClient)
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
			resp, err := parseHTTP11Response("HEAD", writer)
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

	t.Run("upgrade from HTTP/1.0 to HTTP/1.1", func(t *testing.T) {
		defHeaders := map[string]string{
			"Hello": "world",
		}
		request := newRequest()
		request.Protocol = proto.HTTP10
		request.Upgrade = proto.HTTP11
		writer := new(JournalingClient)
		s := getSerializer(defHeaders, request, writer)
		s.Upgrade()

		require.NoError(t, s.Write(proto.HTTP11, http.NewResponse()))

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
		require.Equal(t, "HTTP/1.0", resp.Proto)

		resp, err = stdhttp.ReadResponse(reader, r)
		require.NoError(t, err)
		require.Equal(t, 200, resp.StatusCode)
		require.Equal(t, "HTTP/1.1", resp.Proto)
		require.Contains(t, resp.Header, "Hello")
		require.Equal(t, []string{"world"}, resp.Header["Hello"])
	})

	{
		defHeaders := map[string]string{
			"Hello": "world",
		}
		writer := new(JournalingClient)
		request := newRequest()
		s := getSerializer(defHeaders, request, writer)

		testSized := func(t *testing.T, method string, contentLength int, body string) {
			r, err := parseHTTP11Response(method, writer)
			require.NoError(t, err)
			require.Equal(t, "HTTP/1.1", r.Proto)
			require.Equal(t, "200 OK", r.Status)
			require.Equal(t, []string{"world"}, r.Header["Hello"])
			require.Equal(t, contentLength, int(r.ContentLength))
			require.Empty(t, r.TransferEncoding)
			content, err := io.ReadAll(r.Body)
			require.NoError(t, err)
			require.Equal(t, body, string(content))

			writer.Reset()
		}

		testUnsized := func(t *testing.T, method string, body string) {
			r, err := parseHTTP11Response(method, writer)
			require.NoError(t, err)
			require.Equal(t, "HTTP/1.1", r.Proto)
			require.Equal(t, "200 OK", r.Status)
			require.Equal(t, []string{"world"}, r.Header["Hello"])
			require.Equal(t, -1, int(r.ContentLength))
			require.Equal(t, []string{"chunked"}, r.TransferEncoding)
			content, err := io.ReadAll(r.Body)
			require.NoError(t, err)
			require.Equal(t, body, string(content))

			writer.Reset()
		}

		t.Run("empty", func(t *testing.T) {
			request.Method = method.GET
			resp := http.NewResponse()
			require.NoError(t, s.Write(proto.HTTP11, resp))
			testSized(t, "GET", 0, "")
		})

		const helloworld = "Hello, world!"

		t.Run("plain", func(t *testing.T) {
			request.Method = method.GET
			resp := http.NewResponse().String("Hello, world!")
			require.NoError(t, s.Write(proto.HTTP11, resp))
			testSized(t, "GET", len(helloworld), helloworld)
		})

		t.Run("sized stream", func(t *testing.T) {
			request.Method = method.GET
			resp := http.NewResponse().
				SizedStream(strings.NewReader(helloworld), int64(len(helloworld)))
			require.NoError(t, s.Write(proto.HTTP11, resp))
			testSized(t, "GET", len(helloworld), helloworld)
		})

		t.Run("unsized stream", func(t *testing.T) {
			request.Method = method.GET
			resp := http.NewResponse().
				Stream(strings.NewReader(helloworld))
			require.NoError(t, s.Write(proto.HTTP11, resp))
			testUnsized(t, "GET", helloworld)
		})

		t.Run("HEAD sized stream", func(t *testing.T) {
			request.Method = method.HEAD
			resp := http.NewResponse().
				SizedStream(strings.NewReader(helloworld), int64(len(helloworld)))
			require.NoError(t, s.Write(proto.HTTP11, resp))
			testSized(t, "HEAD", len(helloworld), "")
		})

		t.Run("HEAD unsized stream", func(t *testing.T) {
			request.Method = method.HEAD
			resp := http.NewResponse().
				Stream(strings.NewReader(helloworld))
			require.NoError(t, s.Write(proto.HTTP11, resp))
			testUnsized(t, "HEAD", "")
		})

		t.Run("WriterTo and Closer", func(t *testing.T) {
			request.Method = method.GET
			w := new(writerToCloser)
			w.Reset([]byte(helloworld))
			resp := http.NewResponse().
				SizedStream(w, int64(len(helloworld)))
			require.NoError(t, s.Write(proto.HTTP11, resp))
			testSized(t, "GET", len(helloworld), helloworld)
		})
	}

	t.Run("content-type charset", func(t *testing.T) {
		writer := new(JournalingClient)
		request := newRequest()
		request.Method = method.HEAD
		s := getSerializer(nil, request, writer)

		testMIME := func(t *testing.T, resp *http.Response, wantMIME string) {
			require.NoError(t, s.Write(proto.HTTP11, resp))
			r, err := parseHTTP11Response("HEAD", writer)
			require.NoError(t, err)

			if wantMIME != mime.Unset {
				require.Equal(t, wantMIME, r.Header["Content-Type"][0])
				require.Equal(t, 1, len(r.Header["Content-Type"]))
			} else {
				require.Empty(t, r.Header["Content-Type"])
			}

			writer.Reset()
		}

		t.Run("known default", func(t *testing.T) {
			testMIME(t, http.NewResponse(), "text/html;charset=utf8")
		})

		t.Run("unknown default", func(t *testing.T) {
			testMIME(t, http.NewResponse().ContentType(mime.PNG), "image/png")
		})

		t.Run("explicitly set", func(t *testing.T) {
			testMIME(t, http.NewResponse().ContentType(mime.CSS, mime.UTF32), "text/css;charset=utf32")
		})

		t.Run("no content type", func(t *testing.T) {
			testMIME(t, http.NewResponse().ContentType(mime.Unset), mime.Unset)
		})
	})
}

type writerToCloser struct {
	constReader
	Closed bool
}

func (w *writerToCloser) WriteTo(writer io.Writer) (n int64, err error) {
	for len(w.data) > 0 {
		written, err := writer.Write(w.data)
		w.data = w.data[written:]
		n += int64(written)
		if err != nil {
			return n, err
		}
	}

	return n, nil
}

func (w *writerToCloser) Close() error {
	w.Closed = true
	return nil
}

func TestSerializer_ChunkedTransfer(t *testing.T) {
	t.Run("single chunk", func(t *testing.T) {
		reader := bytes.NewBuffer([]byte("Hello, world!"))
		wantData := "d;0\r\nHello, world!\r\n0;0\r\n\r\n"
		writer := new(JournalingClient)
		s := getSerializer(nil, newRequest(), writer)

		err := s.writeChunked(reader)
		require.NoError(t, err)
		require.Equal(t, wantData, string(writer.Data))
	})

	t.Run("long chunk into small buffer", func(t *testing.T) {
		w := new(JournalingClient)
		const buffSize = 64
		cfg := config.Default()
		cfg.HTTP.ResponseBuffer.Default = buffSize
		cfg.HTTP.ResponseBuffer.Maximal = buffSize
		s := newSerializer(cfg, w, make([]byte, 0, buffSize), nil, newRequest())

		p := newChunkedParser()
		payload := strings.Repeat("Pavlo is the best", 10*buffSize)
		reader := bytes.NewBuffer([]byte(payload))

		require.NoError(t, s.writeChunked(reader))
		out, extra, err := feed(&p, w.Data)
		require.NoError(t, err)
		require.Empty(t, extra)
		require.Equal(t, payload, string(out))
	})
}

func parseHTTP11Response(method string, writer *JournalingClient) (*stdhttp.Response, error) {
	reader := bufio.NewReader(bytes.NewBuffer(writer.Data))
	req, err := stdhttp.NewRequest(method, "/", nil)
	if err != nil {
		return nil, err
	}

	return stdhttp.ReadResponse(reader, req)
}

type dummyConn = dummy.Conn

type JournalingClient struct {
	dummyConn
}

func (j *JournalingClient) Conn() net.Conn {
	return j
}

func (j *JournalingClient) Reset() {
	j.Data = j.Data[:0]
}
