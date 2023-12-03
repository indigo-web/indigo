package inbuilt

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"testing"

	"github.com/indigo-web/indigo/http"

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
		request.Path = "/"
		resp := r.OnRequest(request)
		require.Equal(t, status.OK, resp.Reveal().Code)
	})

	t.Run("POST /", func(t *testing.T) {
		require.Contains(t, r.registrar.routes, "/")
		require.NotNil(t, r.registrar.routes["/"][method.POST])
		request := getRequest()
		request.Method = method.POST
		request.Path = "/"
		resp := r.OnRequest(request)
		require.Equal(t, status.OK, resp.Reveal().Code)
	})

	t.Run("POST /hello", func(t *testing.T) {
		require.Contains(t, r.registrar.routes, "/hello")
		require.NotNil(t, r.registrar.routes["/hello"][method.POST])
		request := getRequest()
		request.Method = method.POST
		request.Path = "/hello"
		resp := r.OnRequest(request)
		require.Equal(t, status.OK, resp.Reveal().Code)
	})

	t.Run("HEAD /", func(t *testing.T) {
		request := getRequest()
		request.Method = method.HEAD
		request.Path = "/"

		resp := r.OnRequest(request)
		// we have not registered any HEAD-method handler yet, so GET method
		// is expected to be called (but without body)
		require.Equal(t, status.OK, resp.Reveal().Code)
		require.Empty(t, string(resp.Reveal().Body))
	})
}

func testMethodShorthand(
	t *testing.T, router *Router,
	route func(string, Handler, ...Middleware) *Router,
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
	request.Path = "/"
	request.Method = method.POST
	response := r.OnRequest(request)
	require.Equal(t, status.MethodNotAllowed, response.Reveal().Code)
}

func TestRouter_RouteError(t *testing.T) {
	r := New()
	r.RouteError(func(req *http.Request) *http.Response {
		return req.Respond().
			Code(status.Teapot).
			String(req.Env.Error.Error())
	}, status.BadRequest)

	t.Run("status.ErrBadRequest", func(t *testing.T) {
		request := getRequest()
		resp := r.OnError(request, status.ErrBadRequest)
		require.Equal(t, status.Teapot, resp.Reveal().Code)
		require.Equal(t, status.ErrBadRequest.Error(), string(resp.Reveal().Body))
	})

	t.Run("status.ErrURIDecoding (also bad request)", func(t *testing.T) {
		request := getRequest()
		resp := r.OnError(request, status.ErrURIDecoding)
		require.Equal(t, status.Teapot, resp.Reveal().Code)
		require.Equal(t, status.ErrURIDecoding.Error(), string(resp.Reveal().Body))
	})

	t.Run("unregistered http error", func(t *testing.T) {
		request := getRequest()
		resp := r.OnError(request, status.ErrNotImplemented)
		require.Equal(t, status.NotImplemented, resp.Reveal().Code)
		require.Equal(t, status.ErrNotImplemented.Error(), string(resp.Reveal().Body))
	})

	t.Run("unregistered ordinary error", func(t *testing.T) {
		request := getRequest()
		resp := r.OnError(request, errors.New("any error text here"))
		require.Equal(t, status.InternalServerError, resp.Reveal().Code)
		require.Empty(t, string(resp.Reveal().Body))
	})

	t.Run("universal handler", func(t *testing.T) {
		const fromUniversal = "from universal handler with love"

		r := New()
		r.RouteError(func(req *http.Request) *http.Response {
			return req.Respond().
				Code(status.Teapot).
				String(fromUniversal)
		}, AllErrors)

		request := getRequest()
		resp := r.OnError(request, status.ErrNotImplemented)
		require.Equal(t, int(status.Teapot), int(resp.Reveal().Code))
		require.Equal(t, fromUniversal, string(resp.Reveal().Body))
	})
}

func TestAliases(t *testing.T) {
	testRootAlias := func(t *testing.T, r *Router) {
		req := getRequest()
		req.Path = "/"
		req.Method = method.GET
		response := r.OnRequest(req)
		require.Equal(t, status.OK, response.Reveal().Code)
		require.Equal(t, "magic word", string(response.Reveal().Body))
	}

	t.Run("single alias", func(t *testing.T) {
		r := New().
			Get("/hello", func(req *http.Request) *http.Response {
				require.Equal(t, "/hello", req.Path)
				require.Equal(t, "/", req.Env.AliasFrom)

				return req.Respond().String("magic word")
			}).
			Alias("/", "/hello")

		require.NoError(t, r.OnStart())

		testRootAlias(t, r)
	})

	t.Run("override normal handler", func(t *testing.T) {
		r := New().
			Get("/", func(req *http.Request) *http.Response {
				require.Fail(t, "the handler must be overridden and never be called")
				return http.Respond(req)
			}).
			Get("/hello", func(req *http.Request) *http.Response {
				require.Equal(t, "/hello", req.Path)
				require.Equal(t, "/", req.Env.AliasFrom)

				return req.Respond().String("magic word")
			}).
			Alias("/", "/hello")

		require.NoError(t, r.OnStart())

		testRootAlias(t, r)
	})

	testOrdinaryRequest := func(t *testing.T, r *Router) {
		req := getRequest()
		req.Path = "/hello"
		req.Method = method.GET
		response := r.OnRequest(req)
		require.Equal(t, status.OK, response.Reveal().Code)
		require.Equal(t, "ordinary word", string(response.Reveal().Body))
	}

	t.Run("multiple calls", func(t *testing.T) {
		var i int

		r := New().
			Get("/", func(req *http.Request) *http.Response {
				require.Fail(t, "the handler must be overridden and never be called")
				return http.Respond(req)
			}).
			Get("/hello", func(req *http.Request) *http.Response {
				defer func() {
					i++
				}()

				if i%2 == 0 {
					require.Equalf(t, "/hello", req.Path, "on iteration: %d", i)
					require.Equalf(t, "/", req.Env.AliasFrom, "on iteration: %d", i)

					return req.Respond().String("magic word")
				} else {
					require.Equalf(t, "/hello", req.Path, "on iteration: %d", i)
					require.Emptyf(t, req.Env.AliasFrom, "on iteration: %d", i)

					return req.Respond().String("ordinary word")
				}
			}).
			Alias("/", "/hello")

		require.NoError(t, r.OnStart())

		testRootAlias(t, r)
		testOrdinaryRequest(t, r)
		testRootAlias(t, r)
		testOrdinaryRequest(t, r)
	})
}

func TestCatchers(t *testing.T) {
	request := func(m method.Method, path string) *http.Request {
		req := getRequest()
		req.Method = m
		req.Path = path
		return req
	}

	t.Run("multiple overriding endpoints", func(t *testing.T) {
		r := New().
			Get("/", http.Respond).
			Get("/hello", http.Respond).
			Catch("/", func(req *http.Request) *http.Response {
				return req.Respond().String("magic")
			}).
			Catch("/hello", func(req *http.Request) *http.Response {
				return req.Respond().String("double magic")
			})
		require.NoError(t, r.OnStart())

		resp := r.OnRequest(request(method.GET, "/"))
		require.Equal(t, status.OK, resp.Reveal().Code)
		require.Empty(t, resp.Reveal().Body)
		resp = r.OnRequest(request(method.GET, "/hello"))
		require.Equal(t, status.OK, resp.Reveal().Code)
		require.Empty(t, resp.Reveal().Body)
		resp = r.OnRequest(request(method.GET, "/anything"))
		require.Equal(t, status.OK, resp.Reveal().Code)
		require.Equal(t, "magic", string(resp.Reveal().Body))
		resp = r.OnRequest(request(method.GET, "/helloworld"))
		require.Equal(t, status.OK, resp.Reveal().Code)
		require.Equal(t, "double magic", string(resp.Reveal().Body))
	})
}
