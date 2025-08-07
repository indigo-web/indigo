package inbuilt

import (
	"testing"

	"github.com/indigo-web/indigo/config"
	"github.com/indigo-web/indigo/http/method"
	"github.com/indigo-web/indigo/http/status"
	"github.com/indigo-web/indigo/internal/construct"
	"github.com/indigo-web/indigo/transport/dummy"
	"github.com/stretchr/testify/require"

	"github.com/indigo-web/indigo/http"
)

/*
This file is separated because it is a bit specific and contains a lot
of specific stuff for testing only middlewares. Decided it's better to
separate it from all the other tests
*/

type middleware uint8

const (
	m1 middleware = iota + 1
	m2
	m3
	m4
	m5
	m6
	m7
)

type callstack struct {
	chain []middleware
}

func (c *callstack) Push(ware middleware) {
	c.chain = append(c.chain, ware)
}

func (c *callstack) Chain() []middleware {
	return c.chain
}

func (c *callstack) Clear() {
	c.chain = c.chain[:0]
}

func getMiddleware(mware middleware, stack *callstack) Middleware {
	return func(next Handler, request *http.Request) *http.Response {
		stack.Push(mware)

		return next(request)
	}
}

func getRequest(m method.Method, path string) *http.Request {
	request := construct.Request(config.Default(), dummy.NewNopClient())
	request.Method = m
	request.Path = path

	return request
}

func TestMiddlewares(t *testing.T) {
	stack := new(callstack)
	raw := New().
		Use(getMiddleware(m1, stack)).
		Get("/", http.Respond, getMiddleware(m2, stack))

	api := raw.Group("/api").
		Use(getMiddleware(m3, stack))

	api.Group("/v1").
		Get("/hello", http.Respond, getMiddleware(m6, stack)).
		Use(getMiddleware(m4, stack))

	api.Group("/v2").
		Use(getMiddleware(m5, stack)).
		Get("/world", http.Respond, getMiddleware(m7, stack))

	r := raw.Build()

	t.Run("/", func(t *testing.T) {
		request := getRequest(method.GET, "/")
		response := r.OnRequest(request)
		require.Equal(t, status.OK, response.Reveal().Code)
		require.Equal(t, []middleware{m1, m2}, stack.Chain())
		stack.Clear()
	})

	t.Run("/api/v1/hello", func(t *testing.T) {
		request := getRequest(method.GET, "/api/v1/hello")
		response := r.OnRequest(request)
		require.Equal(t, status.OK, response.Reveal().Code)
		require.Equal(t, []middleware{m1, m3, m4, m6}, stack.Chain())
		stack.Clear()
	})

	t.Run("/api/v2/world", func(t *testing.T) {
		request := getRequest(method.GET, "/api/v2/world")
		response := r.OnRequest(request)
		require.Equal(t, status.OK, response.Reveal().Code)
		require.Equal(t, []middleware{m1, m3, m5, m7}, stack.Chain())
		stack.Clear()
	})
}
