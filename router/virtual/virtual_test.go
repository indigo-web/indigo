package virtual

import (
	"github.com/indigo-web/indigo/config"
	"github.com/indigo-web/indigo/http"
	"github.com/indigo-web/indigo/http/status"
	"github.com/indigo-web/indigo/internal/construct"
	"github.com/indigo-web/indigo/router/inbuilt"
	"github.com/indigo-web/indigo/transport/dummy"
	"github.com/stretchr/testify/require"
	"testing"
)

func newRequest(hosts ...string) *http.Request {
	request := construct.Request(config.Default(), dummy.NewNopClient())
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
		r := New().Build()
		require.True(t, requestIs(r.OnRequest(newRequest("localhost")), status.MisdirectedRequest))
	})

	t.Run("default router", func(t *testing.T) {
		{
			r := New().
				Default(inbuilt.New()).
				Build()

			require.True(t, requestIs(r.OnRequest(newRequest("localhost")), OK))
			require.True(t, requestIs(r.OnRequest(newRequest("127.0.0.1")), OK))
		}
		{
			r := New().
				Host("0.0.0.0", inbuilt.New()).
				Build()

			require.True(t, requestIs(r.OnRequest(newRequest("localhost")), OK))
			require.True(t, requestIs(r.OnRequest(newRequest("127.0.0.1")), OK))
		}
	})

	t.Run("single host", func(t *testing.T) {
		r := New().
			Host("pavlo.ooo", inbuilt.New()).
			Build()

		require.True(t, requestIs(r.OnRequest(newRequest("pavlo.ooo")), OK))
		require.True(t, requestIs(r.OnRequest(newRequest("localhost")), status.MisdirectedRequest))
		require.True(t, requestIs(r.OnRequest(newRequest("pavlo.ooo", "localhost")), status.BadRequest))
	})
}

func requestIs(resp *http.Response, code status.Code) bool {
	return resp.Reveal().Code == code
}
