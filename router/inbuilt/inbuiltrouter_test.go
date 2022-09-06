package inbuilt

import (
	"testing"

	"github.com/fakefloordiv/indigo/http/status"

	methods "github.com/fakefloordiv/indigo/http/method"
	"github.com/fakefloordiv/indigo/types"

	"github.com/stretchr/testify/require"
)

const respBody = "some body" // once told me

// handler that does nothing, used in cases when we need nothing
// but handler also must not be nil
func nopHandler(_ *types.Request) types.Response {
	return types.WithResponse.WithBody(respBody)
}

func TestRoute(t *testing.T) {
	r := NewRouter()

	t.Run("NewRoute", func(t *testing.T) {
		r.Route(methods.GET, "/", nopHandler)

		require.Contains(t, r.routes, "/")
		require.Equal(t, 1, len(r.routes))
		require.Equal(t, 1, len(r.routes["/"]))
		require.Contains(t, r.routes["/"], methods.GET)
		require.NotNil(t, r.routes["/"][methods.GET])
	})

	t.Run("SamePathDifferentMethod", func(t *testing.T) {
		r.Route(methods.POST, "/", nopHandler)

		require.Contains(t, r.routes, "/")
		require.Equal(t, 1, len(r.routes))
		require.Equal(t, 2, len(r.routes["/"]))
		require.Contains(t, r.routes["/"], methods.POST)
		require.NotNil(t, r.routes["/"][methods.POST])
	})

	t.Run("DifferentPath", func(t *testing.T) {
		r.Route(methods.POST, "/hello", nopHandler)

		require.Contains(t, r.routes, "/hello")
		require.Equal(t, 2, len(r.routes))
		require.Equal(t, 1, len(r.routes["/hello"]))
		require.Contains(t, r.routes["/hello"], methods.POST)
		require.NotNil(t, r.routes["/hello"][methods.POST])
	})

	t.Run("HEAD", func(t *testing.T) {
		request, _ := getRequest()
		request.Method = methods.HEAD
		request.Path = "/"

		resp := r.processRequest(request)
		// we have not registered any HEAD-method handler yet, so GET method
		// is expected to be called (but without body)
		require.Equal(t, status.OK, resp.Code)
		require.Equal(t, respBody, string(resp.Body))
	})
}

func testMethodPredicate(t *testing.T, router *DefaultRouter, route func(string, HandlerFunc, ...Middleware), method methods.Method) {
	route("/", nopHandler)
	require.Contains(t, router.routes, "/")
	require.Contains(t, router.routes["/"], method)
	require.NotNil(t, router.routes["/"][method])
}

func TestMethodPredicates(t *testing.T) {
	r := NewRouter()

	t.Run("GET", func(t *testing.T) {
		testMethodPredicate(t, r, r.Get, methods.GET)
	})
	t.Run("HEAD", func(t *testing.T) {
		testMethodPredicate(t, r, r.Head, methods.HEAD)
	})
	t.Run("POST", func(t *testing.T) {
		testMethodPredicate(t, r, r.Post, methods.POST)
	})
	t.Run("PUT", func(t *testing.T) {
		testMethodPredicate(t, r, r.Put, methods.PUT)
	})
	t.Run("DELETE", func(t *testing.T) {
		testMethodPredicate(t, r, r.Delete, methods.DELETE)
	})
	t.Run("CONNECT", func(t *testing.T) {
		testMethodPredicate(t, r, r.Connect, methods.CONNECT)
	})
	t.Run("OPTIONS", func(t *testing.T) {
		testMethodPredicate(t, r, r.Options, methods.OPTIONS)
	})
	t.Run("TRACE", func(t *testing.T) {
		testMethodPredicate(t, r, r.Trace, methods.TRACE)
	})
	t.Run("PATCH", func(t *testing.T) {
		testMethodPredicate(t, r, r.Patch, methods.PATCH)
	})
}

func TestGroups(t *testing.T) {
	r := NewRouter()

	r.Get("/", nopHandler)

	api := r.Group("/api")

	v1 := api.Group("/v1")
	v1.Get("/hello", nopHandler)

	v2 := api.Group("/v2")
	v2.Get("/world", nopHandler)

	r.OnStart(nil)

	require.Contains(t, r.routes, "/")
	require.Contains(t, r.routes, "/api/v1/hello")
	require.Contains(t, r.routes, "/api/v2/world")
	require.Equal(t, 3, len(r.routes))
}
