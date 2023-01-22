package inbuilt

import (
	"testing"

	"github.com/indigo-web/indigo/internal/server/tcp/dummy"

	"github.com/indigo-web/indigo/http"
	"github.com/indigo-web/indigo/internal/parser/http1"
	"github.com/indigo-web/indigo/settings"

	"github.com/indigo-web/indigo/http/status"

	routertypes "github.com/indigo-web/indigo/router/inbuilt/types"

	"github.com/indigo-web/indigo/http/headers"
	methods "github.com/indigo-web/indigo/http/method"
	"github.com/indigo-web/indigo/http/url"
	"github.com/stretchr/testify/require"
)

/*
This file is separated because it is a bit specific and contains a lot
of specific stuff for testing only middlewares. Decided it's better to
separate it from all the other tests
*/

type middleware uint8

const (
	global1 middleware = iota + 1
	global2
	local1
	local2
	local3
	pointApplied1
	pointApplied2
)

type callstack struct {
	chain []middleware
}

func (c *callstack) Push(ware middleware) {
	c.chain = append(c.chain, ware)
}

func (c callstack) Chain() []middleware {
	return c.chain
}

func (c *callstack) Clear() {
	c.chain = c.chain[:0]
}

func getGlobal1Middleware(stack *callstack) routertypes.Middleware {
	return func(next routertypes.HandlerFunc, request *http.Request) http.Response {
		stack.Push(global1)

		return next(request)
	}
}

func getGlobal2Middleware(stack *callstack) routertypes.Middleware {
	return func(next routertypes.HandlerFunc, request *http.Request) http.Response {
		stack.Push(global2)

		return next(request)
	}
}

func getLocal1Middleware(stack *callstack) routertypes.Middleware {
	return func(next routertypes.HandlerFunc, request *http.Request) http.Response {
		stack.Push(local1)

		return next(request)
	}
}

func getLocal2Middleware(stack *callstack) routertypes.Middleware {
	return func(next routertypes.HandlerFunc, request *http.Request) http.Response {
		stack.Push(local2)

		return next(request)
	}
}

func getLocal3Middleware(stack *callstack) routertypes.Middleware {
	return func(next routertypes.HandlerFunc, request *http.Request) http.Response {
		stack.Push(local3)

		return next(request)
	}
}

func getPointApplied1Middleware(stack *callstack) routertypes.Middleware {
	return func(next routertypes.HandlerFunc, request *http.Request) http.Response {
		stack.Push(pointApplied1)

		return next(request)
	}
}

func getPointApplied2Middleware(stack *callstack) routertypes.Middleware {
	return func(next routertypes.HandlerFunc, request *http.Request) http.Response {
		stack.Push(pointApplied2)

		return next(request)
	}
}

func getRequest() *http.Request {
	query := url.NewQuery(nil)
	bodyReader := http1.NewBodyReader(dummy.NewNopClient(), settings.Default().Body)

	return http.NewRequest(
		headers.NewHeaders(nil), query, http.NewResponse(), dummy.NewNopConn(), bodyReader,
	)
}

func TestMiddlewares(t *testing.T) {
	stack := new(callstack)
	global1mware := getGlobal1Middleware(stack)
	global2mware := getGlobal2Middleware(stack)
	local1mware := getLocal1Middleware(stack)
	local2mware := getLocal2Middleware(stack)
	local3mware := getLocal3Middleware(stack)
	pointApplied1mware := getPointApplied1Middleware(stack)
	pointApplied2mware := getPointApplied2Middleware(stack)

	r := NewRouter()
	r.Use(global1mware)
	r.Get("/", nopHandler, global2mware)

	api := r.Group("/api")
	api.Use(local1mware)

	v1 := api.Group("/v1")
	v1.Use(local2mware)
	v1.Get("/hello", nopHandler, pointApplied1mware)

	v2 := api.Group("/v2")
	v2.Get("/world", nopHandler, pointApplied2mware)
	v2.Use(local3mware)

	r.OnStart()

	t.Run("/", func(t *testing.T) {
		request := getRequest()
		request.Method = methods.GET
		request.Path = "/"

		response := r.OnRequest(request)
		require.Equal(t, int(status.OK), int(response.Code))

		wantChain := []middleware{
			global2, global1,
		}

		require.Equal(t, wantChain, stack.Chain())
		stack.Clear()
	})

	t.Run("/api/v1/hello", func(t *testing.T) {
		request := getRequest()
		request.Method = methods.GET
		request.Path = "/api/v1/hello"

		response := r.OnRequest(request)
		require.Equal(t, status.OK, response.Code)

		wantChain := []middleware{
			pointApplied1, global1, local1, local2,
		}

		require.Equal(t, wantChain, stack.Chain())
		stack.Clear()
	})

	t.Run("/api/v2/world", func(t *testing.T) {
		request := getRequest()
		request.Method = methods.GET
		request.Path = "/api/v2/world"

		response := r.OnRequest(request)
		require.Equal(t, status.OK, response.Code)

		wantChain := []middleware{
			pointApplied2, global1, local1, local3,
		}

		require.Equal(t, wantChain, stack.Chain())
		stack.Clear()
	})
}
