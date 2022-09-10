package inbuilt

import (
	"github.com/fakefloordiv/indigo/types"
)

/*
This file is responsible for middlewares
*/

// Middleware works like a chain of nested calls, next may be even directly
// handler. But if we are not a closing middleware, we will call next
// middleware that is simply a partial middleware with already provided next
type Middleware func(next HandlerFunc, request *types.Request) types.Response

// Use adds a middleware into the global lists of group middlewares. They will
// be applied when registered
func (d *DefaultRouter) Use(middleware Middleware) {
	for _, methods := range d.routes {
		for _, handler := range methods {
			handler.middlewares = append(handler.middlewares, middleware)
		}
	}

	d.middlewares = append(d.middlewares, middleware)
}

func (d Router) applyMiddlewares() {
	for _, methods := range d.routes {
		for _, handler := range methods {
			handler.fun = compose(handler.fun, handler.middlewares)
		}
	}
}

// compose just makes a single HandlerFunc from a chain of middlewares
// and handler in the end using anonymous functions for partials and
// recursion for building a chain (iteration algorithm did not work
// idk why it was causing a recursion)
func compose(handler HandlerFunc, middlewares []Middleware) HandlerFunc {
	if len(middlewares) == 0 {
		return handler
	}

	return func(request *types.Request) types.Response {
		return middlewares[len(middlewares)-1](
			compose(handler, middlewares[:len(middlewares)-1]), request,
		)
	}
}
