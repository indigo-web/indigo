package http1

import (
	"bufio"
	"bytes"
	"fmt"
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
	"github.com/indigo-web/indigo/kv"
	"github.com/indigo-web/indigo/transport/dummy"
	"github.com/stretchr/testify/require"
)

var noCodecs = codecutil.NewCache(nil)

func BenchmarkSerializer(b *testing.B) {
	getRequest := func(cfg *config.Config, m method.Method) *http.Request {
		request := http.NewRequest(cfg, nil, dummy.NewNopClient(), kv.New(), nil, nil)
		request.Method = m
		return request
	}

	getSerializer := func(cfg *config.Config, m method.Method) *serializer {
		buff := make([]byte, 0, cfg.NET.WriteBufferSize.Default)
		return newSerializer(cfg, getRequest(cfg, method.GET), new(dummy.NopClient), noCodecs, buff)
	}

	getResponseWithHeaders := func(n int) *http.Response {
		// in fact there will be 12-13 headers, because Content-Length, Accept-Encoding and
		// eventually Transfer-Encoding aren't counted in

		resp := http.NewResponse()

		for i := range n {
			resp.Header(
				fmt.Sprintf("Test-Header-Key-Of-Index-%d", i+1),
				"test header value of some index",
			)
		}

		return resp
	}

	countResponseSize := func(resp *http.Response) int64 {
		s := getSerializer(config.Default(), method.GET)
		s.client = dummy.NewMockClient().Journaling()
		err := s.Write(proto.HTTP11, resp)
		if err != nil {
			panic(err.Error())
		}

		return int64(len(s.client.(*dummy.Client).Written()))
	}

	runBench := func(cfg *config.Config, m method.Method, resp *http.Response) func(b *testing.B) {
		return func(b *testing.B) {
			s := getSerializer(cfg, m)
			b.SetBytes(countResponseSize(resp))
			b.ReportAllocs()
			b.ResetTimer()

			for range b.N {
				_ = s.Write(proto.HTTP11, resp)
			}
		}
	}

	b.Run("headers", func(b *testing.B) {
		// as headers complexity is estimated to be n*m (where m = default headers = 0),
		// the results must also scale linearly.
		b.Run("10", runBench(config.Default(), method.HEAD, getResponseWithHeaders(10)))
		b.Run("50", runBench(config.Default(), method.HEAD, getResponseWithHeaders(50)))

		cfg1def := config.Default()
		cfg1def.Headers.Default = map[string]string{
			"Server": "indigo",
		}
		b.Run("10+1def", runBench(cfg1def, method.HEAD, getResponseWithHeaders(10)))

		cfg10def := config.Default()
		cfg10def.Headers.Default = map[string]string{
			"Server":                   "indigo",
			"I-gotta-be-creative-here": "only 5 of headers are",
			"allowed-to-collide":       "with user-provided ones",
			"I-hope-that-these":        "headers will shuffle well",
			"However-I-cannot-be":      "a hundred percent sure. Whatever.",
		}
		for _, pair := range getResponseWithHeaders(5).Expose().Headers {
			cfg10def.Headers.Default[pair.Key] = pair.Value
		}

		b.Run("10+10def", runBench(cfg10def, method.HEAD, getResponseWithHeaders(10)))
	})

	b.Run("stream", func(b *testing.B) {
		b.Run("sized 512b", func(b *testing.B) {
			content := strings.Repeat("a", 512)
			resp := http.NewResponse()
			s := getSerializer(config.Default(), method.GET)
			b.SetBytes(512)
			b.ReportAllocs()
			b.ResetTimer()

			for range b.N {
				_ = s.Write(proto.HTTP11, resp.String(content))
			}
		})

		b.Run("unsized 32x16", func(b *testing.B) {
			r := &circularReader{
				n:    32,
				data: []byte(strings.Repeat("a", 16)),
			}
			resp := http.NewResponse()
			s := getSerializer(config.Default(), method.GET)
			b.SetBytes(512)
			b.ReportAllocs()
			b.ResetTimer()

			for range b.N {
				r.n = 32
				_ = s.Write(proto.HTTP11, resp.Stream(r))
			}
		})

		b.Run("unsized 8x64", func(b *testing.B) {
			r := &circularReader{
				n:    8,
				data: []byte(strings.Repeat("a", 64)),
			}
			resp := http.NewResponse()
			s := getSerializer(config.Default(), method.GET)
			b.SetBytes(512)
			b.ReportAllocs()
			b.ResetTimer()

			for range b.N {
				r.n = 8
				_ = s.Write(proto.HTTP11, resp.Stream(r))
			}
		})

		b.Run("sized 262144b", func(b *testing.B) {
			// 262144 = 16 * 16384 = 8 * 32768
			content := strings.Repeat("a", 262144)
			resp := http.NewResponse()
			s := getSerializer(config.Default(), method.GET)
			b.SetBytes(262144)
			b.ReportAllocs()
			b.ResetTimer()

			for range b.N {
				_ = s.Write(proto.HTTP11, resp.String(content))
			}
		})

		b.Run("unsized 16x16384", func(b *testing.B) {
			r := &circularReader{
				n:    16,
				data: []byte(strings.Repeat("a", 16384)),
			}
			resp := http.NewResponse()
			s := getSerializer(config.Default(), method.GET)
			b.SetBytes(262144)
			b.ReportAllocs()
			b.ResetTimer()

			for range b.N {
				r.n = 16
				_ = s.Write(proto.HTTP11, resp.Stream(r))
			}
		})

		b.Run("unsized 8x32768", func(b *testing.B) {
			r := &circularReader{
				n:    8,
				data: []byte(strings.Repeat("a", 32768)),
			}
			resp := http.NewResponse()
			s := getSerializer(config.Default(), method.GET)
			b.SetBytes(262144)
			b.ReportAllocs()
			b.ResetTimer()

			for range b.N {
				r.n = 8
				_ = s.Write(proto.HTTP11, resp.Stream(r))
			}
		})
	})
}

type circularReader struct {
	data, pending []byte
	n             int
}

func (c *circularReader) Read(p []byte) (int, error) {
	if len(c.pending) > 0 {
		n := copy(p, c.pending)
		c.pending = c.pending[n:]
		return n, nil
	}

	if c.n == 0 {
		return 0, io.EOF
	}

	c.n--

	n := copy(p, c.data)
	c.pending = c.data[n:]
	return n, nil
}

func TestCircularReader(t *testing.T) {
	r := &circularReader{data: []byte("Hello"), n: 3}
	buff := make([]byte, 5)
	for range 3 {
		n, err := r.Read(buff)
		require.NoError(t, err)
		require.Equal(t, "Hello", string(buff[:n]))
	}

	_, err := r.Read(buff)
	require.EqualError(t, err, io.EOF.Error())
}

func TestSerializer(t *testing.T) {
	newRequest := func(m method.Method) *http.Request {
		req := construct.Request(config.Default(), dummy.NewNopClient())
		req.Method = m
		return req
	}

	getSerializer := func(defHeaders map[string]string, r *http.Request, codecs codecutil.Cache) (*serializer, *dummy.Client) {
		w := dummy.NewMockClient().Journaling()
		cfg := config.Default()
		cfg.Headers.Default = defHeaders
		buff := make([]byte, 0, config.Default().NET.WriteBufferSize.Default)
		s := newSerializer(cfg, r, w, codecs, buff)
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

			if method == "HEAD" {
				_, b, _ := strings.Cut(string(w.Written()), crlf+crlf)
				require.Empty(t, b)
			}

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

			if method == "HEAD" {
				_, b, _ := strings.Cut(string(w.Written()), crlf+crlf)
				require.Empty(t, b)
			}

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
			resp := http.NewResponse().Stream(strings.NewReader(helloworld), -1)
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
			resp := http.NewResponse().Stream(strings.NewReader(helloworld), -1)
			require.NoError(t, s.Write(proto.HTTP11, resp))
			testUnsized(t, "HEAD", "")
		})

		_ = io.WriterTo(new(strings.Reader))

		t.Run("HEAD WriterTo", func(t *testing.T) {
			w.Reset()
			request.Method = method.HEAD
			resp := http.NewResponse().Stream(strings.NewReader(helloworld))
			require.NoError(t, s.Write(proto.HTTP11, resp))

			testSized(t, "HEAD", len(helloworld), "")
		})

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
			testGZIP(t, http.NewResponse().Stream(strings.NewReader(helloworld), -1))
		})

		t.Run("sized WriterTo", func(t *testing.T) {
			w.Reset()
			request.Method = method.GET
			resp := http.NewResponse().Stream(strings.NewReader(helloworld))
			require.NoError(t, s.Write(proto.HTTP11, resp))

			testSized(t, "GET", len(helloworld), helloworld)
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
	})

	t.Run("writer", func(t *testing.T) {
		t.Run("identity", func(t *testing.T) {
			t.Run("flush", func(t *testing.T) {
				s, w := getSerializer(nil, newRequest(method.GET), noCodecs)
				s.buff = make([]byte, 0, 16)
				writer := identityWriter{s}
				_, err := writer.Write(bytes.Repeat([]byte("a"), 10))
				require.NoError(t, err)
				_, err = writer.Write(bytes.Repeat([]byte("a"), 7))
				require.NoError(t, err)

				require.Equal(t, strings.Repeat("a", 16), string(w.Written()))
				require.Equal(t, "a", string(s.buff))
			})
		})

		t.Run("chunked", func(t *testing.T) {
			init := func(cfg *config.Config, codecs codecutil.Cache) (*serializer, *dummy.Client) {
				client := dummy.NewMockClient().Journaling()
				buff := make([]byte, 0, cfg.NET.WriteBufferSize.Default)
				s := newSerializer(cfg, newRequest(method.GET), client, codecs, buff)

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

			// does essentially the same, just using the ReaderFrom method.
			encodeChunked2 := func(t *testing.T, s *serializer, data ...string) {
				writer := chunkedWriter{s}
				rs := make([]io.Reader, len(data))
				for i, chunk := range data {
					rs[i] = strings.NewReader(chunk)
				}

				_, err := writer.ReadFrom(io.MultiReader(rs...))
				require.NoError(t, err)
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
				want := "2\r\nHe\r\n9\r\nllo, worl\r\n2\r\nd!\r\n0\r\n\r\n"
				require.Equal(t, want, string(w.Written()))
			})

			t.Run("ReaderFrom", func(t *testing.T) {
				s, w := init(config.Default(), noCodecs)
				encodeChunked2(t, s, "Hello, ", "world!")
				want := "7\r\nHello, \r\n6\r\nworld!\r\n0\r\n\r\n"
				require.Equal(t, want, string(w.Written()))
			})
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
}

func parseHTTP11Response(method string, data []byte) (*stdhttp.Response, error) {
	reader := bufio.NewReader(bytes.NewBuffer(data))
	req, err := stdhttp.NewRequest(method, "/", nil)
	if err != nil {
		return nil, err
	}

	return stdhttp.ReadResponse(reader, req)
}
