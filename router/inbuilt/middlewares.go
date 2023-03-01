package inbuilt

import (
	"github.com/indigo-web/indigo/v2/http"

	"github.com/indigo-web/indigo/v2/router/inbuilt/types"
)

/*
This file is responsible for middlewares
*/

// Use adds middlewares into the global list of a group's middlewares. But they will
// be applied only after server will be started
func (r *Router) Use(middlewares ...types.Middleware) {
	for _, methods := range r.routes {
		for _, handler := range methods {
			handler.Middlewares = append(handler.Middlewares, middlewares...)
		}
	}

	r.middlewares = append(r.middlewares, middlewares...)
}

func (r *Router) applyMiddlewares() {
	for _, methods := range r.routes {
		for _, handler := range methods {
			handler.Fun = compose(handler.Fun, handler.Middlewares)
		}
	}
}

// compose just makes a single HandlerFunc from a chain of middlewares
// and handler in the end using anonymous functions for partials and
// recursion for building a chain (iteration algorithm did not work
// IDK why it was causing a recursion)
func compose(handler types.HandlerFunc, middlewares []types.Middleware) types.HandlerFunc {
	if len(middlewares) == 0 {
		return handler
	}

	return func(request *http.Request) http.Response {
		return middlewares[0](
			compose(handler, middlewares[1:]), request,
		)
	}
}
