package http1

import (
	"bufio"
	"bytes"
	"io"
	stdhttp "net/http"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/indigo-web/indigo/config"
	"github.com/indigo-web/indigo/http"
	"github.com/indigo-web/indigo/http/codec"
	"github.com/indigo-web/indigo/http/cookie"
	"github.com/indigo-web/indigo/http/method"
	"github.com/indigo-web/indigo/http/mime"
	"github.com/indigo-web/indigo/http/proto"
	"github.com/indigo-web/indigo/internal/codecutil"
	"github.com/indigo-web/indigo/internal/construct"
	"github.com/indigo-web/indigo/transport/dummy"
	"github.com/stretchr/testify/require"
)

var noCodecs = codecutil.NewCache(nil)

//func BenchmarkSerializer(b *testing.B) {
//	defaultHeadersSmall := map[string]string{
//		"Server": "indigo",
//	}
//	defaultHeadersMedium := map[string]string{
//		"Server":          "indigo",
//		"Connection":      "keep-alive",
//		"Accept-Encoding": "identity",
//	}
//	defaultHeadersBig := map[string]string{
//		"Server":          "indigo",
//		"Connection":      "keep-alive",
//		"Accept-Encoding": "identity",
//		"Easter":          "Egg",
//		"Many":            "choices, variants, ways, solutions",
//		"Something":       "is not happening",
//		"Talking":         "allowed",
//		"Lorem":           "ipsum, doremi",
//	}
//
//	response := http.NewResponse()
//	request := construct.Request(config.Default(), dummy.NewNopClient())
//	w := dummy.NewNopClient()
//
//	b.Run("no body no def headers", func(b *testing.B) {
//		buff := make([]byte, 0, 1024)
//		s := newSerializer(config.Default(), request, w, noCodecs, buff, nil)
//		size, err := estimateResponseSize(request, response, nil)
//		require.NoError(b, err)
//		b.SetBytes(size)
//		b.ReportAllocs()
//		b.ResetTimer()
//
//		for i := 0; i < b.N; i++ {
//			_ = s.Write(request.Protocol, response)
//		}
//	})
//
//	b.Run("with 4kb body", func(b *testing.B) {
//		response := http.NewResponse().String(strings.Repeat("a", 4096))
//		buff := make([]byte, 0, 8192)
//		s := newSerializer(config.Default(), request, w, noCodecs, buff, nil)
//		respSize, err := estimateResponseSize(request, response, nil)
//		require.NoError(b, err)
//		b.SetBytes(respSize)
//		b.ReportAllocs()
//		b.ResetTimer()
//
//		for i := 0; i < b.N; i++ {
//			_ = s.Write(request.Protocol, response)
//		}
//	})
//
//	b.Run("no body 1 def header", func(b *testing.B) {
//		buff := make([]byte, 0, 1024)
//		s := newSerializer(config.Default(), request, w, noCodecs, buff, defaultHeadersSmall)
//		respSize, err := estimateResponseSize(request, response, defaultHeadersSmall)
//		require.NoError(b, err)
//		b.SetBytes(respSize)
//		b.ReportAllocs()
//		b.ResetTimer()
//
//		for i := 0; i < b.N; i++ {
//			_ = s.Write(request.Protocol, response)
//		}
//	})
//
//	b.Run("no body 3 def headers", func(b *testing.B) {
//		buff := make([]byte, 0, 1024)
//		s := newSerializer(config.Default(), request, w, noCodecs, buff, defaultHeadersMedium)
//		respSize, err := estimateResponseSize(request, response, defaultHeadersMedium)
//		require.NoError(b, err)
//		b.SetBytes(respSize)
//		b.ReportAllocs()
//		b.ResetTimer()
//
//		for i := 0; i < b.N; i++ {
//			_ = s.Write(request.Protocol, response)
//		}
//	})
//
//	b.Run("no body 8 def headers", func(b *testing.B) {
//		buff := make([]byte, 0, 1024)
//		s := newSerializer(config.Default(), request, w, noCodecs, buff, defaultHeadersBig)
//		respSize, err := estimateResponseSize(request, response, defaultHeadersBig)
//		require.NoError(b, err)
//		b.SetBytes(respSize)
//		b.ReportAllocs()
//		b.ResetTimer()
//
//		for i := 0; i < b.N; i++ {
//			_ = s.Write(request.Protocol, response)
//		}
//	})
//
//	b.Run("no body 15 headers", func(b *testing.B) {
//		buff := make([]byte, 0, 1024)
//		request := construct.Request(config.Default(), dummy.NewNopClient())
//		request.Headers = generateHeaders(15)
//		s := newSerializer(config.Default(), request, w, noCodecs, buff, nil)
//		size, err := estimateResponseSize(request, response, nil)
//		require.NoError(b, err)
//		b.SetBytes(size)
//		b.ReportAllocs()
//		b.ResetTimer()
//
//		for i := 0; i < b.N; i++ {
//			_ = s.Write(request.Protocol, response)
//		}
//	})
//
//	b.Run("pre-write", func(b *testing.B) {
//		buff := make([]byte, 0, 128)
//		s := newSerializer(config.Default(), request, w, noCodecs, buff, nil)
//		request.Upgrade = proto.HTTP11
//		respSize, err := estimateUpgradeSize(request, response)
//		require.NoError(b, err)
//		b.SetBytes(respSize)
//		b.ReportAllocs()
//		b.ResetTimer()
//
//		for i := 0; i < b.N; i++ {
//			s.Upgrade()
//			_ = s.Write(request.Upgrade, response)
//		}
//	})
//
//	// TODO: add benchmarking chunked body
//}

func estimateResponseSize(req *http.Request, resp *http.Response, defHeaders map[string]string) (int64, error) {
	writer := dummy.NewMockClient()
	s := newSerializer(config.Default(), req, writer, noCodecs, make([]byte, 0, 128), defHeaders)
	err := s.Write(req.Protocol, resp)

	return int64(len(writer.Written())), err
}

func estimateUpgradeSize(
	req *http.Request, resp *http.Response,
) (int64, error) {
	writer := dummy.NewMockClient()
	s := newSerializer(config.Default(), req, writer, noCodecs, make([]byte, 0, 128), nil)
	s.Upgrade()
	err := s.Write(req.Protocol, resp)

	return int64(len(writer.Written())), err
}

func TestSerializer(t *testing.T) {
	newRequest := func(m method.Method) *http.Request {
		req := construct.Request(config.Default(), dummy.NewNopClient())
		req.Method = m
		return req
	}

	getSerializer := func(defHeaders map[string]string, r *http.Request, codecs codecutil.Cache) (*serializer, *dummy.Client) {
		w := dummy.NewMockClient().Journaling()
		buff := make([]byte, 0, config.Default().NET.WriteBufferSize.Default)
		s := newSerializer(config.Default(), r, w, codecs, buff, defHeaders)
		return s, w
	}

	testWithHeaders := func(t *testing.T, s *serializer, writer *dummy.Client) {
		response := http.NewResponse().
			Header("Hello", "world").
			Header("Something", "special", "here")

		require.NoError(t, s.Write(proto.HTTP11, response))
		resp, err := parseHTTP11Response("GET", writer.Written())
		require.NoError(t, err)
		require.Equal(t, 200, resp.StatusCode)

		require.Equal(t, []string{"world"}, resp.Header["Hello"])
		require.Equal(t, []string{"indigo"}, resp.Header["Server"])
		require.Equal(t, []string{"ipsum, something else"}, resp.Header["Lorem"])
		require.Equal(t, []string{"special", "here"}, resp.Header["Something"])

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
		s, writer := getSerializer(defHeaders, newRequest(method.GET), noCodecs)
		// double repeat the test to ensure all default headers are reset correctly
		testWithHeaders(t, s, writer)
		testWithHeaders(t, s, writer)
	})

	parseResp := func(t *testing.T, req *http.Request, resp *http.Response) *stdhttp.Response {
		s, w := getSerializer(nil, req, noCodecs)
		require.NoError(t, s.Write(req.Protocol, resp))

		r, err := parseHTTP11Response(req.Method.String(), w.Written())
		require.NoError(t, err)

		return r
	}

	t.Run("nonstandard code", func(t *testing.T) {
		resp := parseResp(t, newRequest(method.GET), http.NewResponse().Code(600))
		require.Equal(t, "600 Nonstandard", resp.Status)
	})

	t.Run("nonstandard status", func(t *testing.T) {
		resp := parseResp(t, newRequest(method.GET), http.NewResponse().Status("Slava Ukraini!"))
		require.Equal(t, "200 Slava Ukraini!", resp.Status)
	})

	t.Run("unknown request protocol", func(t *testing.T) {
		req := newRequest(method.GET)
		req.Protocol = proto.Unknown
		resp := parseResp(t, req, http.NewResponse())
		require.Equal(t, "HTTP/1.1", resp.Proto)
		require.Equal(t, "200 OK", resp.Status)
	})

	t.Run("upgrade from HTTP/1.0 to HTTP/1.1", func(t *testing.T) {
		defHeaders := map[string]string{
			"Hello": "world",
		}
		request := newRequest(method.GET)
		request.Protocol = proto.HTTP10
		request.Upgrade = proto.HTTP11
		s, w := getSerializer(defHeaders, request, noCodecs)
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
		reader := bufio.NewReader(bytes.NewBuffer(w.Written()))
		resp, err := stdhttp.ReadResponse(reader, r)
		require.NoError(t, err)
		require.Equal(t, 101, resp.StatusCode)
		require.Equal(t, "HTTP/1.0", resp.Proto)

		resp, err = stdhttp.ReadResponse(reader, r)
		require.NoError(t, err)
		require.Equal(t, 200, resp.StatusCode)
		require.Equal(t, "HTTP/1.1", resp.Proto)
		require.Equal(t, []string{"world"}, resp.Header["Hello"])
	})

	t.Run("cookies", func(t *testing.T) {
		t.Run("single pair no params", func(t *testing.T) {
			s, w := getSerializer(nil, newRequest(method.GET), noCodecs)
			response := http.NewResponse().Cookie(cookie.New("hello", "world"))

			require.NoError(t, s.Write(proto.HTTP11, response))
			resp, err := parseHTTP11Response("HEAD", w.Written())
			require.NoError(t, err)
			c := resp.Header.Get("Set-Cookie")
			require.Equal(t, "hello=world", c)
		})

		t.Run("multiple pairs with parameters", func(t *testing.T) {
			s, w := getSerializer(nil, newRequest(method.GET), noCodecs)
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

			require.NoError(t, s.Write(proto.HTTP11, response))
			resp, err := parseHTTP11Response("HEAD", w.Written())
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

	t.Run("streams", func(t *testing.T) {
		request := newRequest(method.GET)
		codecs := codecutil.NewCache([]codec.Codec{codec.NewGZIP()})
		s, w := getSerializer(nil, request, codecs)

		testSized := func(t *testing.T, method string, contentLength int, body string, contentEncoding ...string) {
			r, err := parseHTTP11Response(method, w.Written())
			require.NoError(t, err)
			require.Equal(t, "HTTP/1.1", r.Proto)
			require.Equal(t, "200 OK", r.Status)
			require.Empty(t, r.TransferEncoding)
			require.Equal(t, contentEncoding, r.Header["Content-Encoding"])
			content, err := io.ReadAll(r.Body)
			require.NoError(t, err)
			require.Equal(t, body, string(content))
			require.Equal(t, contentLength, int(r.ContentLength))
		}

		testUnsized := func(t *testing.T, method string, body string, contentEncoding ...string) {
			r, err := parseHTTP11Response(method, w.Written())
			require.NoError(t, err)
			require.Equal(t, "HTTP/1.1", r.Proto)
			require.Equal(t, "200 OK", r.Status)
			require.Equal(t, -1, int(r.ContentLength))
			require.Equal(t, []string{"chunked"}, r.TransferEncoding)
			require.Equal(t, contentEncoding, r.Header["Content-Encoding"])
			content, err := io.ReadAll(r.Body)
			require.NoError(t, err)
			require.Equal(t, body, string(content))
		}

		t.Run("empty", func(t *testing.T) {
			w.Reset()
			request.Method = method.GET
			resp := http.NewResponse()
			require.NoError(t, s.Write(proto.HTTP11, resp))
			testSized(t, "GET", 0, "")
		})

		const helloworld = "Hello, world!"

		t.Run("sized", func(t *testing.T) {
			w.Reset()
			request.Method = method.GET
			resp := http.NewResponse().String(helloworld)
			require.NoError(t, s.Write(proto.HTTP11, resp))
			testSized(t, "GET", len(helloworld), helloworld)
		})

		t.Run("unsized", func(t *testing.T) {
			w.Reset()
			request.Method = method.GET
			resp := http.NewResponse().Stream(strings.NewReader(helloworld))
			require.NoError(t, s.Write(proto.HTTP11, resp))
			testUnsized(t, "GET", helloworld)
		})

		t.Run("HEAD sized", func(t *testing.T) {
			w.Reset()
			request.Method = method.HEAD
			resp := http.NewResponse().String(helloworld)
			require.NoError(t, s.Write(proto.HTTP11, resp))
			testSized(t, "HEAD", len(helloworld), "")
		})

		t.Run("HEAD unsized", func(t *testing.T) {
			w.Reset()
			request.Method = method.HEAD
			resp := http.NewResponse().Stream(strings.NewReader(helloworld))
			require.NoError(t, s.Write(proto.HTTP11, resp))
			testUnsized(t, "HEAD", "")
		})

		// TODO: WriterTo will be implemented in both compressors AND chunked/identity writers. They
		// TODO: will be tested automatically, so...
		//t.Run("WriterTo and Closer", func(t *testing.T) {
		//	w.Reset()
		//	request.Method = method.GET
		//	stream := &closeTracer{readerWriterTo: bytes.NewBuffer([]byte(helloworld))}
		//	resp := http.NewResponse().SizedStream(stream, int64(len(helloworld)))
		//	require.NoError(t, s.Write(proto.HTTP11, resp))
		//	testSized(t, "HEAD", len(helloworld), "")
		//	require.Equal(t, helloworld, string(w.Conn().(*dummy.Conn).Data))
		//	require.True(t, stream.Closed)
		//})

		testGZIP := func(t *testing.T, resp *http.Response) {
			w.Reset()
			request.Method = method.GET

			require.NoError(t, s.Write(proto.HTTP11, resp.Compress("gzip")))
			wantBody := encodeGZIP(helloworld)
			testUnsized(t, "GET", string(wantBody), "gzip")
		}

		t.Run("sized GZIP", func(t *testing.T) {
			testGZIP(t, http.NewResponse().String(helloworld))
		})

		t.Run("unsized GZIP", func(t *testing.T) {
			testGZIP(t, http.NewResponse().Stream(strings.NewReader(helloworld)))
		})

		t.Run("sized buffer growth", func(t *testing.T) {
			writeResp := func(t *testing.T, resp *http.Response, buffsize int, cfg *config.Config) (*serializer, string) {
				s, w := getSerializer(nil, newRequest(method.GET), noCodecs)
				s.buff = make([]byte, 0, buffsize)
				require.NoError(t, s.Write(proto.HTTP11, resp))

				return s, string(w.Written())
			}

			t.Run("fill the buffer exactly full", func(t *testing.T) {
				// estimate headers length, so we can know how many bytes of body we need to trigger the growth
				const buffsize = 128
				s, defaultResponse := writeResp(t, http.NewResponse(), buffsize, config.Default())
				want := strings.Repeat("a", cap(s.buff)-len(defaultResponse)-1)
				s, _ = writeResp(t, http.NewResponse().String(want), buffsize, config.Default())
				require.Equal(t, cap(s.buff), buffsize)
			})

			t.Run("slight growth", func(t *testing.T) {
				const buffsize = 128
				s, defaultResponse := writeResp(t, http.NewResponse(), buffsize, config.Default())
				want := strings.Repeat("a", cap(s.buff)-len(defaultResponse)+1)
				s, _ = writeResp(t, http.NewResponse().String(want), buffsize, config.Default())
				require.Greater(t, cap(s.buff), buffsize)
			})

			t.Run("limit growth by max size", func(t *testing.T) {
				const buffsize = 128
				cfg := config.Default()
				cfg.NET.WriteBufferSize.Maximal = buffsize
				b := strings.Repeat("a", buffsize)
				s, _ := writeResp(t, http.NewResponse().String(b), buffsize-1, cfg)
				wantBuffsize := cap(slices.Grow(make([]byte, buffsize-1), 1))
				require.Equal(t, wantBuffsize, cap(s.buff))
			})
		})

		t.Run("sized buffer flush", func(t *testing.T) {

		})
	})

	t.Run("content-type charset", func(t *testing.T) {
		s, w := getSerializer(nil, newRequest(method.HEAD), noCodecs)

		testMIME := func(t *testing.T, resp *http.Response, wantMIME string) {
			require.NoError(t, s.Write(proto.HTTP11, resp))

			r, err := parseHTTP11Response("HEAD", w.Written())
			require.NoError(t, err)

			if wantMIME != mime.Unset {
				require.Equal(t, 1, len(r.Header["Content-Type"]))
				require.Equal(t, wantMIME, r.Header["Content-Type"][0])
			} else {
				require.Empty(t, r.Header["Content-Type"])
			}

			w.Reset()
		}

		t.Run("without charset", func(t *testing.T) {
			testMIME(t, http.NewResponse().ContentType(mime.PNG), "image/png")
		})

		t.Run("with charset", func(t *testing.T) {
			testMIME(t, http.NewResponse().ContentType(mime.CSS, mime.UTF32), "text/css; charset=utf32")
		})

		t.Run("unset", func(t *testing.T) {
			testMIME(t, http.NewResponse().ContentType(mime.Unset), mime.Unset)
			testMIME(t, http.NewResponse().ContentType(mime.Unset, mime.Unset), mime.Unset)
		})
	})

	t.Run("chunked writer", func(t *testing.T) {
		init := func(cfg *config.Config, codecs codecutil.Cache) (*serializer, *dummy.Client) {
			client := dummy.NewMockClient().Journaling()
			buff := make([]byte, 0, cfg.NET.WriteBufferSize.Default)
			s := newSerializer(cfg, newRequest(method.GET), client, codecs, buff, nil)

			return s, client
		}

		// override the encodeChunked() from suit_test.go because we are interested
		// particularly in the chunkedWriter's output
		encodeChunked := func(t *testing.T, s *serializer, data ...string) {
			writer := chunkedWriter{s}

			for _, chunk := range data {
				_, err := writer.Write([]byte(chunk))
				require.NoError(t, err)
			}
			require.NoError(t, writer.Close())
		}

		t.Run("elide zerofill", func(t *testing.T) {
			s, w := init(config.Default(), noCodecs)
			encodeChunked(t, s, "Hello, ", "world!")
			want := "7\r\nHello, \r\n6\r\nworld!\r\n0\r\n\r\n"
			require.Equal(t, want, string(w.Written()))
		})

		t.Run("use zerofill", func(t *testing.T) {
			s, w := init(config.Default(), noCodecs)
			s.buff = append(s.buff, "Foo! "...)
			encodeChunked(t, s, "Hello, ", "world!")
			want := "Foo! 7;0\r\nHello, \r\n6\r\nworld!\r\n0\r\n\r\n"
			require.Equal(t, want, string(w.Written()))
		})

		t.Run("buffer overflow", func(t *testing.T) {
			const writeBufferSize = 7
			cfg := config.Default()
			cfg.NET.WriteBufferSize.Default = writeBufferSize
			cfg.NET.WriteBufferSize.Maximal = writeBufferSize
			s, w := init(cfg, noCodecs)
			encodeChunked(t, s, "Hello, world!")
			want := "2\r\nHe\r\n2\r\nll\r\n2\r\no,\r\n2\r\n w\r\n2\r\nor\r\n2\r\nld\r\n1\r\n!\r\n0\r\n\r\n"
			require.Equal(t, want, string(w.Written()))
		})

		t.Run("buffer overflow with growth", func(t *testing.T) {
			const writeBufferSize = 7
			cfg := config.Default()
			cfg.NET.WriteBufferSize.Default = writeBufferSize
			s, w := init(cfg, noCodecs)
			encodeChunked(t, s, "Hello, world!")
			want := "2\r\nHe\r\na\r\nllo, world\r\n1\r\n!\r\n0\r\n\r\n"
			require.Equal(t, want, string(w.Written()))
		})
	})
}

func parseHTTP11Response(method string, data []byte) (*stdhttp.Response, error) {
	reader := bufio.NewReader(bytes.NewBuffer(data))
	req, err := stdhttp.NewRequest(method, "/", nil)
	if err != nil {
		return nil, err
	}

	return stdhttp.ReadResponse(reader, req)
}
