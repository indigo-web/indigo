package virtual

import (
	"github.com/indigo-web/indigo/config"
	"github.com/indigo-web/indigo/http"
	"github.com/indigo-web/indigo/http/status"
	"github.com/indigo-web/indigo/internal/initialize"
	"github.com/indigo-web/indigo/internal/server/tcp/dummy"
	"github.com/indigo-web/indigo/router/inbuilt"
	"github.com/stretchr/testify/require"
	"testing"
)

func newRequest(hosts ...string) *http.Request {
	request := initialize.NewRequest(config.Default(), dummy.NewNopConn(), nil)
	for _, host := range hosts {
		request.Headers.Add("Host", host)
	}

	return request
}

func TestVirtualRouter(t *testing.T) {
	// as 404 Not Found can be returned only from an actual router, the value is
	// considered to be positive
	const OK = status.NotFound

	t.Run("no hosts", func(t *testing.T) {
		r := New()
		require.NoError(t, r.OnStart())
		require.True(t, requestIs(r.OnRequest(newRequest("localhost")), status.MisdirectedRequest))
	})

	t.Run("default router", func(t *testing.T) {
		{
			r := New().
				Default(inbuilt.New())

			require.NoError(t, r.OnStart())
			require.True(t, requestIs(r.OnRequest(newRequest("localhost")), OK))
			require.True(t, requestIs(r.OnRequest(newRequest("127.0.0.1")), OK))
		}
		{
			r := New().
				Host("0.0.0.0", inbuilt.New())

			require.NoError(t, r.OnStart())
			require.True(t, requestIs(r.OnRequest(newRequest("localhost")), OK))
			require.True(t, requestIs(r.OnRequest(newRequest("127.0.0.1")), OK))
		}
	})

	t.Run("single host", func(t *testing.T) {
		r := New().
			Host("pavlo.gay", inbuilt.New())

		require.NoError(t, r.OnStart())
		require.True(t, requestIs(r.OnRequest(newRequest("pavlo.gay")), OK))
		require.True(t, requestIs(r.OnRequest(newRequest("localhost")), status.MisdirectedRequest))
		require.True(t, requestIs(r.OnRequest(newRequest("pavlo.gay", "localhost")), status.BadRequest))
	})
}

func requestIs(resp *http.Response, code status.Code) bool {
	return resp.Reveal().Code == code
}
