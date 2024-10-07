package http

import (
	"github.com/indigo-web/indigo/config"
	"github.com/indigo-web/indigo/http/cookie"
	"github.com/indigo-web/indigo/http/headers"
	"github.com/indigo-web/indigo/http/query"
	"github.com/indigo-web/indigo/internal/tcp/dummy"
	"github.com/stretchr/testify/require"
	"testing"
)

func newRequest() *Request {
	return NewRequest(
		config.Default(), headers.New(), query.New(nil), nil, dummy.NewNopClient(),
		NewBody(nil, config.Default()), nil,
	)
}

func TestCookies(t *testing.T) {
	t.Run("no cookies", func(t *testing.T) {
		request := newRequest()
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

		request := newRequest()
		request.Headers.Add("Cookie", "hello=world; world=hello")
		request.Headers.Add("Cookie", "monke=funny")
		// repeat the test twice to make sure, that calling it again won't produce
		// different result
		test(t, request)
		test(t, request)
	})

	t.Run("malformed", func(t *testing.T) {
		request := newRequest()
		request.Headers.Add("Cookie", "a")
		// repeat the test twice to make sure, that calling it again won't produce
		// different result
		_, err := request.Cookies()
		require.EqualError(t, err, cookie.ErrBadCookie.Error())
		_, err = request.Cookies()
		require.EqualError(t, err, cookie.ErrBadCookie.Error())
	})
}
