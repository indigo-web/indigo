package inbuilt

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/indigo-web/indigo/config"
	"github.com/indigo-web/indigo/internal/construct"
	"github.com/indigo-web/indigo/router"
	"github.com/indigo-web/indigo/transport/dummy"
	"github.com/stretchr/testify/assert"

	"github.com/indigo-web/indigo/http"

	"github.com/indigo-web/indigo/http/status"

	"github.com/indigo-web/indigo/http/method"
	"github.com/stretchr/testify/require"
)

func readbody(t *testing.T, r io.Reader) string {
	data, err := io.ReadAll(r)
	require.NoError(t, err)
	return string(data)
}

func BenchmarkRouter_OnRequest_Static(b *testing.B) {
	raw := New()

	GETRootRequest := getRequest(method.GET, "/")
	raw.Get(GETRootRequest.Path, http.Respond)
	longURIRequest := getRequest(method.GET, "/"+strings.Repeat("a", 65534))
	raw.Get(longURIRequest.Path, http.Respond)
	mediumURIRequest := getRequest(method.GET, "/"+strings.Repeat("a", 50))
	raw.Get(mediumURIRequest.Path, http.Respond)
	unknownURIRequest := getRequest(method.GET, "/"+strings.Repeat("b", 65534))
	unknownMethodRequest := getRequest(method.POST, longURIRequest.Path)

	emptyCtx := context.Background()

	r := raw.Build()

	b.Run("GET root", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			r.OnRequest(GETRootRequest)
			GETRootRequest.Ctx = emptyCtx
		}
	})

	b.Run("GET long uri", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			r.OnRequest(longURIRequest)
			longURIRequest.Ctx = emptyCtx
		}
	})

	b.Run("GET medium uri", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			r.OnRequest(mediumURIRequest)
			mediumURIRequest.Ctx = emptyCtx
		}
	})

	b.Run("unknown uri", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			r.OnRequest(unknownURIRequest)
			unknownURIRequest.Ctx = emptyCtx
		}
	})

	b.Run("unknown method", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			r.OnRequest(unknownMethodRequest)
			unknownMethodRequest.Ctx = emptyCtx
		}
	})
}

func TestRoute(t *testing.T) {
	raw := New()
	raw.Route(method.GET, "/", http.Respond)
	raw.Route(method.POST, "/", http.Respond)
	raw.Route(method.POST, "/hello", http.Respond)
	r := raw.Build()

	t.Run("GET /", func(t *testing.T) {
		request := getRequest(method.GET, "/")
		resp := r.OnRequest(request)
		require.Equal(t, status.OK, resp.Expose().Code)
	})

	t.Run("POST /", func(t *testing.T) {
		request := getRequest(method.POST, "/")
		resp := r.OnRequest(request)
		require.Equal(t, status.OK, resp.Expose().Code)
	})

	t.Run("POST /hello", func(t *testing.T) {
		request := getRequest(method.POST, "/hello")
		resp := r.OnRequest(request)
		require.Equal(t, status.OK, resp.Expose().Code)
	})

	t.Run("HEAD /", func(t *testing.T) {
		request := getRequest(method.HEAD, "/")
		resp := r.OnRequest(request)
		// we have not registered any HEAD-method handler yet, so GET method
		// is expected to be called (but without body)
		require.Equal(t, status.OK, resp.Expose().Code)
		require.Nil(t, resp.Expose().Stream)
	})
}

func TestDynamic(t *testing.T) {
	t.Run("first level", func(t *testing.T) {
		raw := New().
			Get("/:name", func(request *http.Request) *http.Response {
				return http.String(request, request.Vars.Value("name"))
			})
		r := raw.Build()

		request := getRequest(method.GET, "/hello")
		resp := r.OnRequest(request)
		require.Equal(t, "hello", readbody(t, resp.Expose().Stream))
	})

	t.Run("second level", func(t *testing.T) {
		raw := New().
			Get("/hello/:name", func(request *http.Request) *http.Response {
				return http.String(request, request.Vars.Value("name"))
			})
		r := raw.Build()

		request := getRequest(method.GET, "/hello/pavlo")
		resp := r.OnRequest(request)
		require.Equal(t, "pavlo", readbody(t, resp.Expose().Stream))
	})

	t.Run("in the middle", func(t *testing.T) {
		r := New().
			Get("/api/:method/doc", func(request *http.Request) *http.Response {
				return http.String(request, request.Vars.Value("method"))
			}).
			Build()

		request := getRequest(method.GET, "/api/getUser/doc")
		resp := r.OnRequest(request)
		require.Equal(t, "getUser", readbody(t, resp.Expose().Stream))
	})

	t.Run("anonymous section", func(t *testing.T) {
		r := New().
			Get("/:", func(request *http.Request) *http.Response {
				return http.String(request, "yay")
			}).
			Build()

		request := getRequest(method.GET, "/api")
		resp := r.OnRequest(request)
		require.Equal(t, "yay", readbody(t, resp.Expose().Stream))
		request = getRequest(method.GET, "/api/second-level")
		resp = r.OnRequest(request)
		require.Equal(t, int(status.NotFound), int(resp.Expose().Code))
	})
}

func testMethodShorthand(
	t *testing.T, router *Router,
	route func(string, Handler, ...Middleware) *Router,
	method method.Method,
) {
	route("/", http.Respond)
	require.Contains(t, router.registrar.endpoints, "/")
	require.NotNil(t, router.registrar.endpoints["/"][method])
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
	raw := New().
		Get("/", http.Respond)

	api := raw.Group("/api")

	api.Group("/v1").
		Get("/hello", http.Respond)

	api.Group("/v2").
		Get("/world", http.Respond)

	r := raw.Build()

	require.Equal(t, status.OK, r.OnRequest(getRequest(method.GET, "/")).Expose().Code)
	require.Equal(t, status.OK, r.OnRequest(getRequest(method.GET, "/api/v1/hello")).Expose().Code)
	require.Equal(t, status.OK, r.OnRequest(getRequest(method.GET, "/api/v2/world")).Expose().Code)
}

func TestResource(t *testing.T) {
	raw := New()
	raw.Resource("/").
		Get(http.Respond).
		Post(http.Respond)

	api := raw.Group("/api")
	api.Resource("/stat").
		Get(http.Respond).
		Post(http.Respond)

	r := raw.Build()

	t.Run("Root", func(t *testing.T) {
		require.Equal(t, status.OK, r.OnRequest(getRequest(method.GET, "/")).Expose().Code)
		require.Equal(t, status.OK, r.OnRequest(getRequest(method.POST, "/")).Expose().Code)
	})

	t.Run("Group", func(t *testing.T) {
		require.Equal(t, status.OK, r.OnRequest(getRequest(method.GET, "/api/stat")).Expose().Code)
		require.Equal(t, status.OK, r.OnRequest(getRequest(method.POST, "/api/stat")).Expose().Code)
	})
}

func TestResource_Methods(t *testing.T) {
	echoMethod := func(req *http.Request) *http.Response {
		return req.Respond().Status(req.Method.String())
	}

	raw := New()
	raw.Resource("/").
		Get(echoMethod).
		Head(echoMethod).
		Post(echoMethod).
		Put(echoMethod).
		Delete(echoMethod).
		Connect(echoMethod).
		Options(echoMethod).
		Trace(echoMethod).
		Patch(echoMethod).
		Mkcol(echoMethod).
		Move(echoMethod).
		Copy(echoMethod).
		Lock(echoMethod).
		Unlock(echoMethod).
		Propfind(echoMethod).
		Proppatch(echoMethod)

	r := raw.Build()

	for _, m := range method.List {
		resp := r.OnRequest(getRequest(m, "/")).Expose()
		if assert.Equal(t, int(status.OK), int(resp.Code)) {
			assert.Equal(t, m.String(), resp.Status)
		}
	}
}

func TestRouter_MethodNotAllowed(t *testing.T) {
	r := New().
		Get("/", http.Respond).
		Build()

	request := getRequest(method.POST, "/")
	response := r.OnRequest(request)
	require.Equal(t, status.MethodNotAllowed, response.Expose().Code)
}

func TestRouter_RouteError(t *testing.T) {
	r := New().
		RouteError(func(req *http.Request) *http.Response {
			return req.Respond().
				Code(status.Teapot).
				String(req.Env.Error.Error())
		}, status.BadRequest).
		Build()

	t.Run("status.ErrBadRequest", func(t *testing.T) {
		request := getRequest(method.GET, "/")
		resp := r.OnError(request, status.ErrBadRequest)
		require.Equal(t, status.Teapot, resp.Expose().Code)
		require.Equal(t, status.ErrBadRequest.Error(), readbody(t, resp.Expose().Stream))
	})

	t.Run("status.ErrURIDecoding (also bad request)", func(t *testing.T) {
		request := getRequest(method.GET, "/")
		resp := r.OnError(request, status.ErrURLDecoding)
		require.Equal(t, status.Teapot, resp.Expose().Code)
		require.Equal(t, status.ErrURLDecoding.Error(), readbody(t, resp.Expose().Stream))
	})

	t.Run("unregistered http error", func(t *testing.T) {
		request := getRequest(method.GET, "/")
		resp := r.OnError(request, status.ErrNotImplemented)
		require.Equal(t, status.NotImplemented, resp.Expose().Code)
	})

	t.Run("unregistered ordinary error", func(t *testing.T) {
		request := getRequest(method.GET, "/")
		resp := r.OnError(request, errors.New("any error text here"))
		require.Equal(t, status.InternalServerError, resp.Expose().Code)
		require.Nil(t, resp.Expose().Stream)
	})

	t.Run("universal handler", func(t *testing.T) {
		const sample = "from universal handler with love"

		r := New().
			RouteError(func(req *http.Request) *http.Response {
				return req.Respond().
					Code(status.Teapot).
					String(sample)
			}, AllErrors).
			Build()

		request := getRequest(method.GET, "/")
		resp := r.OnError(request, status.ErrNotImplemented)
		require.Equal(t, int(status.Teapot), int(resp.Expose().Code))
		data, err := io.ReadAll(resp.Expose().Stream)
		require.NoError(t, err)
		require.Equal(t, sample, string(data))
	})
}

func TestAliases(t *testing.T) {
	testRootAlias := func(t *testing.T, r router.Router) {
		request := getRequest(method.GET, "/")
		response := r.OnRequest(request)
		require.Equal(t, status.OK, response.Expose().Code)
		data, err := io.ReadAll(response.Expose().Stream)
		require.NoError(t, err)
		require.Equal(t, "magic word", string(data))
	}

	t.Run("single alias", func(t *testing.T) {
		r := New().
			Get("/hello", func(req *http.Request) *http.Response {
				require.Equal(t, "/hello", req.Path)
				require.Equal(t, "/", req.Env.AliasFrom)

				return req.Respond().String("magic word")
			}).
			Alias("/", "/hello")

		testRootAlias(t, r.Build())
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

		testRootAlias(t, r.Build())
	})

	testOrdinaryRequest := func(t *testing.T, r router.Router) {
		request := getRequest(method.GET, "/hello")
		response := r.OnRequest(request)
		require.Equal(t, status.OK, response.Expose().Code)
		data, err := io.ReadAll(response.Expose().Stream)
		require.NoError(t, err)
		require.Equal(t, "ordinary word", string(data))
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

		testRootAlias(t, r.Build())
		testOrdinaryRequest(t, r.Build())
		testRootAlias(t, r.Build())
		testOrdinaryRequest(t, r.Build())
	})

	t.Run("aliases on groups", func(t *testing.T) {
		r := New().
			Get("/heaven", http.Respond)

		r.Group("/hello").
			Alias("/world", "/heaven")

		request := getRequest(method.GET, "/hello/world")
		response := r.Build().OnRequest(request)
		require.Equal(t, int(status.OK), int(response.Expose().Code))
	})
}

func TestMutators(t *testing.T) {
	var timesCalled int

	r := New().
		Get("/", http.Respond).
		Mutator(func(request *http.Request) {
			timesCalled++
		}).
		Build()

	request := construct.Request(config.Default(), dummy.NewNopClient())
	request.Method = method.GET
	request.Path = "/"
	resp := r.OnRequest(request)
	require.Equal(t, status.OK, resp.Expose().Code)

	request.Method = method.POST
	request.Path = "/"
	resp = r.OnRequest(request)
	require.Equal(t, status.MethodNotAllowed, resp.Expose().Code)

	request.Method = method.GET
	request.Path = "/foo"
	resp = r.OnRequest(request)
	require.Equal(t, status.NotFound, resp.Expose().Code)

	require.Equal(t, 3, timesCalled)
}
