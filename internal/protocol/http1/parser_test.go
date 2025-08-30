package http1

import (
	"errors"
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/dchest/uniuri"
	"github.com/indigo-web/indigo/config"
	"github.com/indigo-web/indigo/http"
	"github.com/indigo-web/indigo/http/method"
	"github.com/indigo-web/indigo/http/proto"
	"github.com/indigo-web/indigo/http/status"
	"github.com/indigo-web/indigo/internal/construct"
	"github.com/indigo-web/indigo/kv"
	"github.com/indigo-web/indigo/transport/dummy"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func getParser(cfg *config.Config) (*Parser, *http.Request) {
	request := construct.Request(cfg, dummy.NewNopClient())
	statusLine, headers := construct.Buffers(cfg)
	p := NewParser(cfg, request, statusLine, headers)

	return p, request
}

func BenchmarkParser(b *testing.B) {
	parser, request := getParser(config.Default())

	b.Run("with 5 headers", func(b *testing.B) {
		data := generateRequest(strings.Repeat("a", 500), generateHeaders(5))
		b.SetBytes(int64(len(data)))
		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			_, _, _ = parser.Parse(data)
			request.Reset()
		}
	})

	b.Run("with 10 headers", func(b *testing.B) {
		data := generateRequest(strings.Repeat("a", 500), generateHeaders(10))
		b.SetBytes(int64(len(data)))
		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			_, _, _ = parser.Parse(data)
			request.Reset()
		}
	})

	b.Run("with 50 headers", func(b *testing.B) {
		data := generateRequest(strings.Repeat("a", 500), generateHeaders(50))
		b.SetBytes(int64(len(data)))
		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			_, _, _ = parser.Parse(data)
			request.Reset()
		}
	})

	b.Run("escaped 10 headers", func(b *testing.B) {
		data := generateRequest(strings.Repeat("%20", 500), generateHeaders(10))
		b.SetBytes(int64(len(data)))
		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			_, _, _ = parser.Parse(data)
			request.Reset()
		}
	})
}

type wantedRequest struct {
	Headers  http.Headers
	Path     string
	Method   method.Method
	Protocol proto.Protocol
}

func compareRequests(t *testing.T, wanted wantedRequest, actual *http.Request) {
	require.Equal(t, wanted.Method, actual.Method)
	require.Equal(t, wanted.Path, actual.Path)
	require.Equal(t, wanted.Protocol, actual.Protocol)

	for key := range wanted.Headers.Keys() {
		wh, ah := wanted.Headers.Values(key), actual.Headers.Values(key)
		require.Equal(t, slices.Collect(wh), slices.Collect(ah))
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

func feedPartially(p *Parser, raw []byte, n int) (done bool, extra []byte, err error) {
	parts := splitIntoParts(raw, n)

	for i, chunk := range parts {
		done, extra, err = p.Parse(chunk)
		if err != nil {
			return done, extra, err
		}
		if done {
			if i+1 < len(parts) {
				return true, extra, errors.New("not all chunks were fed")
			}

			break
		}
	}

	return done, extra, err
}

func TestParser(t *testing.T) {
	cfg := config.Default()
	cfg.Headers.MaxEncodingTokens = 3

	t.Run("simple GET", func(t *testing.T) {
		raw := "GET / HTTP/1.1\r\n\r\n"
		parser, request := getParser(cfg)
		done, extra, err := parser.Parse([]byte(raw))
		require.NoError(t, err)
		require.True(t, done)
		require.Empty(t, extra)

		wanted := wantedRequest{
			Method:   method.GET,
			Path:     "/",
			Protocol: proto.HTTP11,
			Headers:  kv.New(),
		}

		compareRequests(t, wanted, request)
	})

	t.Run("leading CRLF", func(t *testing.T) {
		raw := "\r\n\r\nGET / HTTP/1.1\r\n\r\n"
		parser, _ := getParser(cfg)
		done, extra, err := parser.Parse([]byte(raw))
		// unfortunately, we don't support this. Such clients must die.
		require.Error(t, err, status.ErrBadRequest.Error())
		require.True(t, done)
		require.Empty(t, extra)
	})

	t.Run("GET with headers", func(t *testing.T) {
		raw := "GET / HTTP/1.1\r\nHello: World!\r\nEaster: Egg\r\n\r\n"
		parser, request := getParser(cfg)
		done, extra, err := parser.Parse([]byte(raw))
		require.NoError(t, err)
		require.True(t, done)
		require.Empty(t, extra)

		wanted := wantedRequest{
			Method:   method.GET,
			Path:     "/",
			Protocol: proto.HTTP11,
			Headers: kv.NewFromMap(map[string][]string{
				"hello": {"World!"},
			}),
		}

		compareRequests(t, wanted, request)
	})

	t.Run("multiple header values", func(t *testing.T) {
		raw := "GET / HTTP/1.1\r\nAccept: one,two\r\nAccept: three\r\n\r\n"
		parser, request := getParser(cfg)
		done, extra, err := parser.Parse([]byte(raw))
		require.NoError(t, err)
		require.True(t, done)
		require.Empty(t, extra)

		wanted := wantedRequest{
			Method:   method.GET,
			Path:     "/",
			Protocol: proto.HTTP11,
			Headers: kv.NewFromMap(map[string][]string{
				"accept": {"one,two", "three"},
			}),
		}

		compareRequests(t, wanted, request)
	})

	t.Run("only lf", func(t *testing.T) {
		raw := "GET / HTTP/1.1\nHello: World!\n\n"
		parser, request := getParser(cfg)
		done, extra, err := parser.Parse([]byte(raw))
		require.NoError(t, err)
		require.True(t, done)
		require.Empty(t, extra)

		wanted := wantedRequest{
			Method:   method.GET,
			Path:     "/",
			Protocol: proto.HTTP11,
			Headers: kv.NewFromMap(map[string][]string{
				"hello": {"World!"},
			}),
		}

		compareRequests(t, wanted, request)
	})

	t.Run("fuzz GET", func(t *testing.T) {
		raw := "GET / HTTP/1.1\r\nHello: World!\r\nEaster: Egg\r\n\r\n"
		parser, request := getParser(cfg)

		for i := 1; i < len(raw); i++ {
			done, extra, err := feedPartially(parser, []byte(raw), i)
			require.NoError(t, err, i)
			require.Empty(t, extra)
			require.True(t, done)

			wanted := wantedRequest{
				Method:   method.GET,
				Path:     "/",
				Protocol: proto.HTTP11,
				Headers: kv.NewFromMap(map[string][]string{
					"hello": {"World!"},
				}),
			}

			compareRequests(t, wanted, request)
			request.Reset()
		}
	})

	t.Run("absolute path", func(t *testing.T) {
		raw := "GET http://www.w3.org/pub/WWW/TheProject.html HTTP/1.1\r\n\r\n"
		parser, request := getParser(cfg)
		done, extra, err := parser.Parse([]byte(raw))
		require.NoError(t, err)
		require.True(t, done)
		require.Empty(t, extra)

		wanted := wantedRequest{
			Method:   method.GET,
			Path:     "http://www.w3.org/pub/WWW/TheProject.html",
			Protocol: proto.HTTP11,
			Headers:  kv.New(),
		}

		compareRequests(t, wanted, request)
	})

	t.Run("content length", func(t *testing.T) {
		raw := "GET / HTTP/1.1\r\nContent-Length: 13\n\r\nHello, world!"
		parser, request := getParser(cfg)
		done, extra, err := parser.Parse([]byte(raw))
		require.NoError(t, err)
		require.True(t, done)
		require.Equal(t, "Hello, world!", string(extra))
		require.Equal(t, 13, request.ContentLength)
		require.Equal(t, "13", request.Headers.Value("content-length"))
		request.Reset()

		raw = "GET / HTTP/1.1\r\nContent-Length: 13\r\nHi-Hi: ha-ha\r\n\r\nHello, world!"
		done, extra, err = parser.Parse([]byte(raw))
		require.NoError(t, err)
		require.True(t, done)
		require.Equal(t, "Hello, world!", string(extra))
		require.Equal(t, 13, request.ContentLength)
		require.Equal(t, "13", request.Headers.Value("content-length"))
		require.Equal(t, "ha-ha", request.Headers.Value("hi-hi"))
	})

	t.Run("connection", func(t *testing.T) {
		raw := "GET / HTTP/1.1\r\nConnection: Keep-Alive\r\n\r\n"
		parser, request := getParser(cfg)
		done, extra, err := parser.Parse([]byte(raw))
		require.NoError(t, err)
		require.True(t, done)
		require.Empty(t, string(extra))
		require.Equal(t, "Keep-Alive", request.Connection)
	})

	t.Run("Transfer-Encoding and Content-Encoding", func(t *testing.T) {
		raw := "GET / HTTP/1.1\r\nTransfer-Encoding: chunked\r\nContent-Encoding: gzip, deflate\r\n\r\n"
		parser, request := getParser(cfg)
		done, extra, err := parser.Parse([]byte(raw))
		require.NoError(t, err)
		require.True(t, done)
		require.Empty(t, string(extra))
		require.Equal(t, []string{"chunked"}, request.TransferEncoding)
		require.True(t, request.Chunked)
		require.Equal(t, []string{"gzip", "deflate"}, request.ContentEncoding)
	})

	t.Run("urldecode", func(t *testing.T) {
		parseRequestLine := func(path string) (*http.Request, error) {
			raw := fmt.Sprintf("GET %s ", path)
			parser, request := getParser(cfg)

			for i := 0; i < len(raw); i++ {
				_, _, err := parser.Parse([]byte{raw[i]})
				if err != nil {
					return request, err
				}
			}

			return request, nil
		}

		parsePath := func(path string) (string, error) {
			request, err := parseRequestLine(path)
			return request.Path, err
		}

		t.Run("path", func(t *testing.T) {
			path, err := parsePath("/%41%41%41")
			require.NoError(t, err)
			require.Equal(t, "/AAA", path)
		})

		t.Run("path urlencoded unicode", func(t *testing.T) {
			t.Run("fastpath", func(t *testing.T) {
				parser, request := getParser(config.Default())
				_, _, err := parser.Parse([]byte("GET /%D0%9F%D0%B0%D0%B2%D0%BB%D0%BE "))
				require.NoError(t, err)
				require.Equal(t, "/%d0%9f%d0%b0%d0%b2%d0%bb%d0%be", request.Path)
			})

			t.Run("slowpath", func(t *testing.T) {
				request, err := parseRequestLine("/%D0%9F%D0%B0%D0%B2%D0%BB%D0%BE")
				require.NoError(t, err)
				require.Equal(t, "/%d0%9f%d0%b0%d0%b2%d0%bb%d0%be", request.Path)
			})
		})

		t.Run("unsafe", func(t *testing.T) {
			test := func(t *testing.T, request *http.Request) {
				require.Equal(t, "/ foo%2f:bar#?", request.Path)
				require.Equal(t, "bar", request.Params.Value("foo+="))
			}

			t.Run("slowpath", func(t *testing.T) {
				request, err := parseRequestLine("/%20foo%2f%3abar%23%3f?foo%2b%3d=bar")
				require.NoError(t, err)
				test(t, request)
			})

			t.Run("fastpath", func(t *testing.T) {
				parser, request := getParser(config.Default())
				_, _, err := parser.Parse([]byte("GET /%20foo%2f%3abar%23%3f?foo%2b%3d=bar "))
				require.NoError(t, err)
				test(t, request)
			})

			t.Run("normalize", func(t *testing.T) {
				parser, request := getParser(config.Default())
				_, _, err := parser.Parse([]byte("GET /foo%2Fbar "))
				require.NoError(t, err)
				require.Equal(t, "/foo%2fbar", request.Path)
			})
		})

		t.Run("params", func(t *testing.T) {
			parseParams := func(params ...string) (http.Params, error) {
				parser, request := getParser(config.Default())
				_, _, err := parser.Parse([]byte("GET /?"))
				if err != nil {
					return nil, err
				}

				if request.Path != "/" {
					panic("assert: bad request path")
				}

				for _, param := range params {
					_, _, err = parser.Parse([]byte(param))
					if err != nil {
						return nil, err
					}
				}

				_, _, err = parser.Parse([]byte(" "))
				return request.Params, err
			}

			t.Run("fastpath", func(t *testing.T) {
				params, err := parseParams("hello%20world=Slava+%55kraini")
				require.NoError(t, err)
				require.Equal(t, "Slava Ukraini", params.Value("hello world"))
			})

			splitchars := func(str string) []string {
				byChars := make([]string, len(str))
				for i := range str {
					byChars[i] = string(str[i])
				}

				return byChars
			}

			t.Run("slowpath", func(t *testing.T) {
				str := "%53lava+%55kraini=%48eroyam+%53lava"
				params, err := parseParams(splitchars(str)...)
				require.NoError(t, err)
				fmt.Println(params.Expose())
				require.Equal(t, "Heroyam Slava", params.Value("Slava Ukraini"))
			})

			t.Run("allow unicode values", func(t *testing.T) {
				t.Run("fastpath", func(t *testing.T) {
					parser, request := getParser(config.Default())
					_, _, err := parser.Parse([]byte("GET /?n=Слава+Україні "))
					require.NoError(t, err)
					require.Equal(t, "Слава Україні", request.Params.Value("n"))
				})

				t.Run("slowpath", func(t *testing.T) {
					request, err := parseRequestLine("/?n=Слава+Україні")
					require.NoError(t, err)
					require.Equal(t, "Слава Україні", request.Params.Value("n"))
				})
			})
		})
	})

	t.Run("edgecase", func(t *testing.T) {
		const buffsize = 64
		cfg := config.Default()
		cfg.URI.RequestLineSize.Default = buffsize
		cfg.URI.RequestLineSize.Maximal = buffsize

		t.Run("method", func(t *testing.T) {
			for _, tc := range []struct {
				Name, Method string
				WantError    error
			}{
				{
					Name:      "absent",
					Method:    "",
					WantError: status.ErrBadRequest,
				},
				{
					Name:      "short unrecognized",
					Method:    "GE",
					WantError: status.ErrMethodNotImplemented,
				},
				{
					Name:      "standard unrecognized",
					Method:    "GOT",
					WantError: status.ErrMethodNotImplemented,
				},
				{
					Name:      "long unrecognized",
					Method:    "PROPPATCHCOMPAREANDSWAP",
					WantError: status.ErrMethodNotImplemented,
				},
				{
					Name:      "too long unrecognized",
					Method:    strings.Repeat("A", buffsize),
					WantError: status.ErrMethodNotImplemented,
				},
			} {
				parser, _ := getParser(cfg)
				request := fmt.Sprintf("%s / HTTP/1.1\r\n\r\n", tc.Method)
				done, _, err := parser.Parse([]byte(request))
				require.True(t, done)
				require.EqualError(t, err, tc.WantError.Error())
			}

			t.Run("too long split", func(t *testing.T) {
				m := strings.Repeat("A", buffsize)
				parser, _ := getParser(cfg)

				for _, char := range m {
					done, _, err := parser.Parse([]byte{byte(char)})
					require.NoError(t, err)
					require.False(t, done)
				}

				done, _, err := parser.Parse([]byte("A"))
				require.EqualError(t, err, status.ErrMethodNotImplemented.Error())
				require.True(t, done)
			})
		})

		t.Run("path", func(t *testing.T) {
			parser, _ := getParser(cfg)
			_, _, err := parser.Parse([]byte("GET / HTTP/1.1\r\n"))
			require.NoError(t, err)

			pathlimit := buffsize - parser.requestLine.Len() + 1

			for _, tc := range []struct {
				Name, Sample string
				WantError    error
			}{
				{
					Name:      "absent",
					Sample:    "",
					WantError: status.ErrBadRequest,
				},
				{
					Name:      "single whitespace",
					Sample:    " ",
					WantError: status.ErrBadRequest,
				},
				{
					Name:      "too long",
					Sample:    "/" + strings.Repeat("a", pathlimit),
					WantError: status.ErrURITooLong,
				},
				{
					Name:      "too long percent-separated",
					Sample:    "/" + strings.Repeat("a", pathlimit) + "%2a",
					WantError: status.ErrURITooLong,
				},
				{
					Name:      "too long question mark-separated",
					Sample:    "/" + strings.Repeat("a", pathlimit) + "?",
					WantError: status.ErrURITooLong,
				},
				{
					Name:      "plain nonprintable",
					Sample:    "/\x00",
					WantError: status.ErrBadRequest,
				},
				{
					Name:      "encoded nonprintable",
					Sample:    "/%ff",
					WantError: nil,
				},
				{
					Name:      "param key nonprintable",
					Sample:    "/?he%07llo=world",
					WantError: status.ErrBadParams,
				},
				{
					Name:      "fragment",
					Sample:    "/hello#Section1",
					WantError: status.ErrBadRequest,
				},
				{
					Name:      "fragment after query",
					Sample:    "/?hello=world#Section1",
					WantError: status.ErrBadRequest,
				},
				{
					Name:      "fragment after query flag",
					Sample:    "/?hello#Section1",
					WantError: status.ErrBadRequest,
				},
				{
					Name:      "OK length",
					Sample:    "/" + strings.Repeat("%2a", pathlimit-1),
					WantError: nil,
				},
				{
					Name:      "too long encoded",
					Sample:    "/" + strings.Repeat("%2a", pathlimit),
					WantError: status.ErrURITooLong,
				},
			} {
				t.Run(tc.Name, func(t *testing.T) {
					parser, _ = getParser(cfg)
					request := fmt.Sprintf("GET %s HTTP/1.1\r\n\r\n", tc.Sample)
					done, _, err := parser.Parse([]byte(request))
					require.True(t, done)

					if tc.WantError == nil {
						require.NoError(t, err)
						return
					}

					require.EqualError(t, err, tc.WantError.Error())
				})
			}
		})

		t.Run("protocol", func(t *testing.T) {
			for _, tc := range []struct {
				Name, Proto string
			}{
				{
					Name:  "short invalid",
					Proto: "HTT",
				},
				{
					Name:  "long invalid",
					Proto: "HTTPS/1.1",
				},
				{
					Name:  "unsupported minor",
					Proto: "HTTP/1.2",
				},
				{
					Name:  "unsupported major",
					Proto: "HTTP/42.0",
				},
				{
					Name:  "invalid minor",
					Proto: "HTTP/1.X",
				},
				{
					Name:  "invalid major",
					Proto: "HTTP/X.1",
				},
			} {
				t.Run(tc.Name, func(t *testing.T) {
					parser, _ := getParser(config.Default())
					request := fmt.Sprintf("GET / %s\r\n\r\n", tc.Proto)
					done, _, err := parser.Parse([]byte(request))
					require.True(t, done)
					require.EqualError(t, err, status.ErrHTTPVersionNotSupported.Error())
				})
			}
		})

		t.Run("lfcr crlf break sequence", func(t *testing.T) {
			parser, _ := getParser(config.Default())
			raw := []byte("GET / HTTP/1.1\n\r\r\n")
			done, _, err := parser.Parse(raw)
			require.EqualError(t, err, status.ErrBadRequest.Error())
			require.True(t, done)
		})

		t.Run("lfcr lfcr break sequence", func(t *testing.T) {
			// our parser is able to parse both crlf and lf splitters
			// so in example below he sees LF CRLF CR
			// the last one CR will be returned as extra-bytes
			parser, _ := getParser(config.Default())
			raw := []byte("GET / HTTP/1.1\n\r\n\r")
			done, extra, err := parser.Parse(raw)
			require.Equal(t, []byte("\r"), extra)
			require.NoError(t, err)
			require.True(t, done)
		})

		t.Run("invalid content length", func(t *testing.T) {
			parser, _ := getParser(config.Default())
			raw := []byte("GET / HTTP/1.1\r\nContent-Length: 1f5\r\n\r\n")
			_, _, err := parser.Parse(raw)
			require.EqualError(t, err, status.ErrBadRequest.Error())
		})

		t.Run("simple request", func(t *testing.T) {
			// Simple Requests are not supported, because our server is
			// HTTP/1.1-oriented, and in 1.1 simple request/response is
			// something like a deprecated mechanism
			parser, _ := getParser(config.Default())
			raw := []byte("GET / \r\n")
			done, _, err := parser.Parse(raw)
			require.EqualError(t, err, status.ErrHTTPVersionNotSupported.Error())
			require.True(t, done)
		})

		t.Run("too many headers", func(t *testing.T) {
			parser, _ := getParser(config.Default())
			hdrs := genHeaders(int(config.Default().Headers.Number.Maximal + 1))
			raw := fmt.Sprintf(
				"GET / HTTP/1.1\r\n%s\r\n\r\n",
				strings.Join(hdrs, "\r\n"),
			)
			_, _, err := parser.Parse([]byte(raw))
			require.EqualError(t, err, status.ErrTooManyHeaders.Error())
		})

		t.Run("too many Transfer-Encoding tokens", func(t *testing.T) {
			cfg := config.Default()
			cfg.Headers.MaxEncodingTokens = 3
			parser, _ := getParser(cfg)
			raw := "GET / HTTP/1.1\r\nTransfer-Encoding: gzip, deflate, br, chunked\r\n\r\n"
			done, extra, err := parser.Parse([]byte(raw))
			require.True(t, done)
			require.Empty(t, string(extra))
			require.EqualError(t, err, status.ErrTooManyEncodingTokens.Error())
		})

		t.Run("duplicate Transfer-Encoding", func(t *testing.T) {
			raw := "GET / HTTP/1.1\r\nTransfer-Encoding: gzip\r\nTransfer-Encoding: chunked\r\n\r\n"
			parser, request := getParser(config.Default())
			_, _, err := parser.Parse([]byte(raw))
			require.EqualError(t, err, status.ErrBadEncoding.Error())
			request.Reset()
		})
	})
}

func TestSplitTokens(t *testing.T) {
	t.Run("positive", func(t *testing.T) {
		for i, tc := range []struct {
			Sample string
			Want   []string
		}{
			{"", []string{}},
			{"identity", []string{}},
			{"chunked", []string{"chunked"}},
			{"chunked,gzip", []string{"chunked", "gzip"}},
			{"gzip,chunked", []string{"gzip", "chunked"}},
			{" gzip,    chunked  ", []string{"gzip", "chunked"}},
		} {
			_, toks, err := splitTokens(make([]string, 0, 2), tc.Sample)
			if assert.NoError(t, err) {
				assert.Equal(t, tc.Want, toks, i+1)
			}
		}
	})

	t.Run("negative", func(t *testing.T) {
		for i, tc := range []string{
			", chunked",
			"chunked, ",
			"gzip,,chunked",
			"gzip, , chunked",
		} {
			_, _, err := splitTokens(make([]string, 0, 2), tc)
			assert.EqualError(t, err, status.ErrUnsupportedEncoding.Error(), i+1)
		}
	})

	t.Run("overflow tokens limit", func(t *testing.T) {
		_, toks, err := splitTokens(make([]string, 0, 1), "gzip,flate,chunked")
		require.EqualError(t, err, status.ErrTooManyEncodingTokens.Error())
		require.Nil(t, toks)
	})
}

func genHeaders(n int) (out []string) {
	for i := 0; i < n; i++ {
		out = append(out, genHeader())
	}

	return out
}

func genHeader() string {
	return fmt.Sprintf("%[1]s: %[1]s", uniuri.NewLen(16))
}
