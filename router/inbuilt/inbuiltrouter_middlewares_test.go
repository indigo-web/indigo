package inbuilt

import (
	"context"
	"github.com/fakefloordiv/indigo/http/status"
	"testing"

	routertypes "github.com/fakefloordiv/indigo/router/inbuilt/types"

	"github.com/fakefloordiv/indigo/internal/body"

	"github.com/fakefloordiv/indigo/http/headers"
	methods "github.com/fakefloordiv/indigo/http/method"
	"github.com/fakefloordiv/indigo/http/url"
	"github.com/fakefloordiv/indigo/settings"
	"github.com/fakefloordiv/indigo/types"

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
	return func(ctx context.Context, next routertypes.HandlerFunc, request *types.Request) types.Response {
		stack.Push(global1)

		return next(ctx, request)
	}
}

func getGlobal2Middleware(stack *callstack) routertypes.Middleware {
	return func(ctx context.Context, next routertypes.HandlerFunc, request *types.Request) types.Response {
		stack.Push(global2)

		return next(ctx, request)
	}
}

func getLocal1Middleware(stack *callstack) routertypes.Middleware {
	return func(ctx context.Context, next routertypes.HandlerFunc, request *types.Request) types.Response {
		stack.Push(local1)

		return next(ctx, request)
	}
}

func getLocal2Middleware(stack *callstack) routertypes.Middleware {
	return func(ctx context.Context, next routertypes.HandlerFunc, request *types.Request) types.Response {
		stack.Push(local2)

		return next(ctx, request)
	}
}

func getLocal3Middleware(stack *callstack) routertypes.Middleware {
	return func(ctx context.Context, next routertypes.HandlerFunc, request *types.Request) types.Response {
		stack.Push(local3)

		return next(ctx, request)
	}
}

func getPointApplied1Middleware(stack *callstack) routertypes.Middleware {
	return func(ctx context.Context, next routertypes.HandlerFunc, request *types.Request) types.Response {
		stack.Push(pointApplied1)

		return next(ctx, request)
	}
}

func getPointApplied2Middleware(stack *callstack) routertypes.Middleware {
	return func(ctx context.Context, next routertypes.HandlerFunc, request *types.Request) types.Response {
		stack.Push(pointApplied2)

		return next(ctx, request)
	}
}

func getRequest() (*types.Request, *body.Gateway) {
	manager := headers.NewManager(settings.Default().Headers)
	query := url.NewQuery(nil)

	return types.NewRequest(&manager, query, nil)
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
		request, _ := getRequest()
		request.Method = methods.GET
		request.Path = "/"

		response := r.OnRequest(request)
		require.Equal(t, status.OK, response.Code)

		wantChain := []middleware{
			global1, global2,
		}

		require.Equal(t, wantChain, stack.Chain())
		stack.Clear()
	})

	t.Run("/api/v1/hello", func(t *testing.T) {
		request, _ := getRequest()
		request.Method = methods.GET
		request.Path = "/api/v1/hello"

		response := r.OnRequest(request)
		require.Equal(t, status.OK, response.Code)

		wantChain := []middleware{
			local2, local1, global1, pointApplied1,
		}

		require.Equal(t, wantChain, stack.Chain())
		stack.Clear()
	})

	t.Run("/api/v2/world", func(t *testing.T) {
		request, _ := getRequest()
		request.Method = methods.GET
		request.Path = "/api/v2/world"

		response := r.OnRequest(request)
		require.Equal(t, status.OK, response.Code)

		wantChain := []middleware{
			local3, local1, global1, pointApplied2,
		}

		require.Equal(t, wantChain, stack.Chain())
		stack.Clear()
	})
}
