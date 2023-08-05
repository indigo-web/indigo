package inbuilt

import (
	"github.com/stretchr/testify/assert"
	"testing"

	"github.com/indigo-web/indigo/http"

	"github.com/indigo-web/indigo/router/inbuilt/types"

	"github.com/indigo-web/indigo/http/status"

	"github.com/indigo-web/indigo/http/method"
	"github.com/stretchr/testify/require"
)

func TestRoute(t *testing.T) {
	r := New()
	r.Route(method.GET, "/", http.Respond)
	r.Route(method.POST, "/", http.Respond)
	r.Route(method.POST, "/hello", http.Respond)
	require.NoError(t, r.OnStart())

	t.Run("GET /", func(t *testing.T) {
		require.Contains(t, r.registrar.routes, "/")
		require.NotNil(t, r.registrar.routes["/"][method.GET])
		request := getRequest()
		request.Method = method.GET
		request.Path.String = "/"
		resp := r.OnRequest(request)
		require.Equal(t, status.OK, resp.Code)
	})

	t.Run("POST /", func(t *testing.T) {
		require.Contains(t, r.registrar.routes, "/")
		require.NotNil(t, r.registrar.routes["/"][method.POST])
		request := getRequest()
		request.Method = method.POST
		request.Path.String = "/"
		resp := r.OnRequest(request)
		require.Equal(t, status.OK, resp.Code)
	})

	t.Run("POST /hello", func(t *testing.T) {
		require.Contains(t, r.registrar.routes, "/hello")
		require.NotNil(t, r.registrar.routes["/hello"][method.POST])
		request := getRequest()
		request.Method = method.POST
		request.Path.String = "/hello"
		resp := r.OnRequest(request)
		require.Equal(t, status.OK, resp.Code)
	})

	t.Run("HEAD /", func(t *testing.T) {
		request := getRequest()
		request.Method = method.HEAD
		request.Path.String = "/"

		resp := r.OnRequest(request)
		// we have not registered any HEAD-method handler yet, so GET method
		// is expected to be called (but without body)
		require.Equal(t, status.OK, resp.Code)
		require.Empty(t, string(resp.Body))
	})
}

func testMethodShorthand(
	t *testing.T, router *Router,
	route func(string, types.Handler, ...types.Middleware) *Router,
	method method.Method,
) {
	route("/", http.Respond)
	require.Contains(t, router.registrar.routes, "/")
	require.NotNil(t, router.registrar.routes["/"][method])
}

func TestMethodShorthands(t *testing.T) {
	r := New()

	t.Run("GET", func(t *testing.T) {
		testMethodShorthand(t, r, r.Get, method.GET)
	})
	t.Run("HEAD", func(t *testing.T) {
		testMethodShorthand(t, r, r.Head, method.HEAD)
	})
	t.Run("POST", func(t *testing.T) {
		testMethodShorthand(t, r, r.Post, method.POST)
	})
	t.Run("PUT", func(t *testing.T) {
		testMethodShorthand(t, r, r.Put, method.PUT)
	})
	t.Run("DELETE", func(t *testing.T) {
		testMethodShorthand(t, r, r.Delete, method.DELETE)
	})
	t.Run("CONNECT", func(t *testing.T) {
		testMethodShorthand(t, r, r.Connect, method.CONNECT)
	})
	t.Run("OPTIONS", func(t *testing.T) {
		testMethodShorthand(t, r, r.Options, method.OPTIONS)
	})
	t.Run("TRACE", func(t *testing.T) {
		testMethodShorthand(t, r, r.Trace, method.TRACE)
	})
	t.Run("PATCH", func(t *testing.T) {
		testMethodShorthand(t, r, r.Patch, method.PATCH)
	})
}

func TestGroups(t *testing.T) {
	r := New().
		Get("/", http.Respond)

	api := r.Group("/api")

	api.Group("/v1").
		Get("/hello", http.Respond)

	api.Group("/v2").
		Get("/world", http.Respond)

	require.NoError(t, r.OnStart())

	require.Contains(t, r.registrar.routes, "/")
	require.Contains(t, r.registrar.routes, "/api/v1/hello")
	require.Contains(t, r.registrar.routes, "/api/v2/world")
	require.Equal(t, 3, len(r.registrar.routes))
}

func TestResource(t *testing.T) {
	r := New()
	r.Resource("/").
		Get(http.Respond).
		Post(http.Respond)

	api := r.Group("/api")
	api.Resource("/stat").
		Get(http.Respond).
		Post(http.Respond)

	require.NoError(t, r.OnStart())

	t.Run("Root", func(t *testing.T) {
		require.Contains(t, r.registrar.routes, "/")
		rootMethods := r.registrar.routes["/"]
		require.NotNil(t, rootMethods[method.GET])
		require.NotNil(t, rootMethods[method.POST])
	})

	t.Run("Group", func(t *testing.T) {
		require.Contains(t, r.registrar.routes, "/api/stat")
		apiMethods := r.registrar.routes["/api/stat"]
		require.NotNil(t, apiMethods[method.GET])
		require.NotNil(t, apiMethods[method.POST])
	})
}

func TestResource_Methods(t *testing.T) {
	r := New()
	r.Resource("/").
		Get(http.Respond).
		Head(http.Respond).
		Post(http.Respond).
		Put(http.Respond).
		Delete(http.Respond).
		Connect(http.Respond).
		Options(http.Respond).
		Trace(http.Respond).
		Patch(http.Respond)
	require.NoError(t, r.OnStart())
	require.Contains(t, r.registrar.routes, "/")

	for _, handlerObject := range r.registrar.routes["/"] {
		assert.NotNil(t, handlerObject)
	}
}

func TestRouter_MethodNotAllowed(t *testing.T) {
	r := New().
		Get("/", http.Respond)
	require.NoError(t, r.OnStart())

	request := getRequest()
	request.Path.String = "/"
	request.Method = method.POST
	response := r.OnRequest(request)
	require.Equal(t, status.MethodNotAllowed, response.Code)
}
