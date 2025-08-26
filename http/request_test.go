package http

import (
	"testing"

	"github.com/indigo-web/indigo/config"
	"github.com/indigo-web/indigo/http/cookie"
	"github.com/indigo-web/indigo/kv"
	"github.com/indigo-web/indigo/transport/dummy"
	"github.com/stretchr/testify/require"
)

func getRequest() *Request {
	return NewRequest(config.Default(), nil, dummy.NewNopClient(), kv.New(), kv.New(), kv.New())
}

func TestRequest(t *testing.T) {
	t.Run("cookies", func(t *testing.T) {
		t.Run("none", func(t *testing.T) {
			request := getRequest()
			jar, err := request.Cookies()
			require.NoError(t, err)
			require.Zero(t, jar.Len())
		})

		t.Run("happy path", func(t *testing.T) {
			test := func(t *testing.T, request *Request) {
				jar, err := request.Cookies()
				require.NoError(t, err)
				require.Equal(t, "world", jar.Value("hello"))
				require.Equal(t, "hello", jar.Value("world"))
				require.Equal(t, "funny", jar.Value("monke"))
				require.Equal(t, 3, jar.Len(), "jar must contain exactly 3 pairs")
			}

			request := getRequest()
			request.Headers.Add("Cookie", "hello=world; world=hello")
			request.Headers.Add("Cookie", "monke=funny")
			// repeat the test twice to make sure, that calling it again won't produce
			// different result
			test(t, request)
			test(t, request)
		})

		t.Run("malformed", func(t *testing.T) {
			request := getRequest()
			request.Headers.Add("Cookie", "a")
			// repeat the test twice to make sure, that calling it again won't produce
			// different result
			_, err := request.Cookies()
			require.EqualError(t, err, cookie.ErrBadCookie.Error())
			_, err = request.Cookies()
			require.EqualError(t, err, cookie.ErrBadCookie.Error())
		})
	})

	t.Run("preferred encoding", func(t *testing.T) {
		testPreferredEncoding := func(want string, tokens ...string) func(t *testing.T) {
			return func(t *testing.T) {
				headers := &commonHeaders{AcceptEncoding: tokens}
				require.Equal(t, want, headers.PreferredEncoding())
			}
		}

		t.Run("accept none", testPreferredEncoding("identity"))
		t.Run("no qualifiers", testPreferredEncoding("gzip", "gzip", "deflate"))
		t.Run("with qualifiers", testPreferredEncoding(
			"zstd",
			"gzip", "deflate;q=0.5", "zstd;q=1.0",
		))
		t.Run("invalid qualifier", testPreferredEncoding("gzip", "gzip", "zstd;q=0.05"))
	})
}
