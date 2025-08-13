package inbuilt

import (
	"io"
	"testing"

	"github.com/indigo-web/indigo/config"
	"github.com/indigo-web/indigo/http"
	"github.com/indigo-web/indigo/http/method"
	"github.com/indigo-web/indigo/http/status"
	"github.com/indigo-web/indigo/internal/construct"
	"github.com/indigo-web/indigo/kv"
	"github.com/indigo-web/indigo/transport/dummy"
	"github.com/stretchr/testify/require"
)

func TestTrace(t *testing.T) {
	newRequest := func(path string, params http.Params) *http.Request {
		req := construct.Request(config.Default(), dummy.NewNopClient())
		req.Method = method.TRACE
		req.Path = path
		req.Params = params
		req.Headers = kv.New().
			Add("Accept", "*/*").
			Add("Content-Length", "13")

		return req
	}

	wantMirroring := func(path string) string {
		return "TRACE " + path + " HTTP/1.1\r\nAccept: */*\r\nContent-Length: 13\r\n\r\n"
	}

	r := New().
		EnableTRACE(true).
		Get("/", http.Respond).
		Build()

	t.Run("TRACE on registered endpoint", func(t *testing.T) {
		req := newRequest("/", kv.New())
		resp := r.OnRequest(req)
		require.Equal(t, 200, int(resp.Expose().Code))
		b, err := io.ReadAll(resp.Expose().Stream)
		require.NoError(t, err)
		require.Equal(t, wantMirroring("/"), string(b))
	})

	t.Run("TRACE on non-existing endpoint", func(t *testing.T) {
		req := newRequest("/hello", kv.New())
		resp := r.OnRequest(req)
		require.Equal(t, 200, int(resp.Expose().Code))
		b, err := io.ReadAll(resp.Expose().Stream)
		require.NoError(t, err)
		require.Equal(t, wantMirroring("/hello"), string(b))
	})

	t.Run("params", func(t *testing.T) {
		t.Run("single", func(t *testing.T) {
			req := newRequest("/", kv.New().Add("hello", "world"))
			resp := r.OnRequest(req)
			require.Equal(t, 200, int(resp.Expose().Code))
			b, err := io.ReadAll(resp.Expose().Stream)
			require.NoError(t, err)
			require.Equal(t, wantMirroring("/?hello=world"), string(b))
		})

		t.Run("multiple", func(t *testing.T) {
			params := kv.New().
				Add("hello", "world").
				Add("hi", "hello")
			req := newRequest("/", params)
			resp := r.OnRequest(req)
			require.Equal(t, 200, int(resp.Expose().Code))
			b, err := io.ReadAll(resp.Expose().Stream)
			require.NoError(t, err)
			require.Equal(t, wantMirroring("/?hello=world&hi=hello"), string(b))
		})
	})

	t.Run("disabled", func(t *testing.T) {
		r := New().
			Get("/", http.Respond).
			Build()
		resp := r.OnRequest(newRequest("/", kv.New()))
		require.Equal(t, int(status.MethodNotAllowed), int(resp.Expose().Code))
	})
}
