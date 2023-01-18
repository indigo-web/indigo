package inbuilt

import (
	"testing"

	"github.com/fakefloordiv/indigo/http"

	routertypes "github.com/fakefloordiv/indigo/router/inbuilt/types"

	"github.com/fakefloordiv/indigo/http/status"

	methods "github.com/fakefloordiv/indigo/http/method"
	"github.com/stretchr/testify/require"
)

// handler that does nothing, used in cases when we need nothing
// but handler also must not be nil
func nopHandler(request *http.Request) http.Response {
	return http.RespondTo(request)
}

func TestRoute(t *testing.T) {
	r := NewRouter()
	r.OnStart()

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
		request := getRequest()
		request.Method = methods.HEAD
		request.Path = "/"

		resp := r.processRequest(request)
		// we have not registered any HEAD-method handler yet, so GET method
		// is expected to be called (but without body)
		require.Equal(t, status.OK, resp.Code)
		require.Empty(t, string(resp.Body))
	})
}

func testMethodShorthand(
	t *testing.T, router *Router,
	route func(string, routertypes.HandlerFunc, ...routertypes.Middleware),
	method methods.Method,
) {
	route("/", nopHandler)
	require.Contains(t, router.routes, "/")
	require.Contains(t, router.routes["/"], method)
	require.NotNil(t, router.routes["/"][method])
}

func TestMethodShorthands(t *testing.T) {
	r := NewRouter()

	t.Run("GET", func(t *testing.T) {
		testMethodShorthand(t, r, r.Get, methods.GET)
	})
	t.Run("HEAD", func(t *testing.T) {
		testMethodShorthand(t, r, r.Head, methods.HEAD)
	})
	t.Run("POST", func(t *testing.T) {
		testMethodShorthand(t, r, r.Post, methods.POST)
	})
	t.Run("PUT", func(t *testing.T) {
		testMethodShorthand(t, r, r.Put, methods.PUT)
	})
	t.Run("DELETE", func(t *testing.T) {
		testMethodShorthand(t, r, r.Delete, methods.DELETE)
	})
	t.Run("CONNECT", func(t *testing.T) {
		testMethodShorthand(t, r, r.Connect, methods.CONNECT)
	})
	t.Run("OPTIONS", func(t *testing.T) {
		testMethodShorthand(t, r, r.Options, methods.OPTIONS)
	})
	t.Run("TRACE", func(t *testing.T) {
		testMethodShorthand(t, r, r.Trace, methods.TRACE)
	})
	t.Run("PATCH", func(t *testing.T) {
		testMethodShorthand(t, r, r.Patch, methods.PATCH)
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

	r.OnStart()

	require.Contains(t, r.routes, "/")
	require.Contains(t, r.routes, "/api/v1/hello")
	require.Contains(t, r.routes, "/api/v2/world")
	require.Equal(t, 3, len(r.routes))
}

func TestResource(t *testing.T) {
	r := NewRouter()
	root := r.Resource("/")
	root.Get(nopHandler)
	root.Post(nopHandler)

	api := r.Group("/api")
	stat := api.Resource("/stat")
	stat.Get(nopHandler)
	stat.Post(nopHandler)

	r.OnStart()

	t.Run("Root", func(t *testing.T) {
		require.Contains(t, r.routes, "/")
		rootMethods := r.routes["/"]
		require.Contains(t, rootMethods, methods.GET)
		require.Contains(t, rootMethods, methods.POST)
		require.Equal(
			t, 2, len(rootMethods),
			"only GET and POST methods are expected to be presented",
		)
	})

	t.Run("Group", func(t *testing.T) {
		require.Contains(t, r.routes, "/api/stat")
		apiMethods := r.routes["/api/stat"]
		require.Contains(t, apiMethods, methods.GET)
		require.Contains(t, apiMethods, methods.POST)
		require.Equal(
			t, 2, len(apiMethods),
			"only GET and POST methods are expected to be presented",
		)
	})
}

func TestResource_Methods(t *testing.T) {
	r := NewRouter()
	root := r.Resource("/")
	root.Get(nopHandler)
	root.Head(nopHandler)
	root.Post(nopHandler)
	root.Put(nopHandler)
	root.Delete(nopHandler)
	root.Connect(nopHandler)
	root.Options(nopHandler)
	root.Trace(nopHandler)
	root.Patch(nopHandler)

	require.Contains(t, r.routes, "/")
	require.Equal(t, 9, len(r.routes["/"]))
}
