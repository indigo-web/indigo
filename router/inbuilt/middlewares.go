package inbuilt

import (
	"context"

	routertypes "github.com/fakefloordiv/indigo/router/inbuilt/types"
	"github.com/fakefloordiv/indigo/types"
)

/*
This file is responsible for middlewares
*/

// Use adds a middleware into the global lists of group middlewares. They will
// be applied when registered
func (r *Router) Use(middleware routertypes.Middleware) {
	for _, methods := range r.routes {
		for _, handler := range methods {
			handler.Middlewares = append(handler.Middlewares, middleware)
		}
	}

	r.middlewares = append(r.middlewares, middleware)
}

func (r Router) applyMiddlewares() {
	for _, methods := range r.routes {
		for _, handler := range methods {
			handler.Fun = compose(handler.Fun, handler.Middlewares)
		}
	}
}

// compose just makes a single HandlerFunc from a chain of middlewares
// and handler in the end using anonymous functions for partials and
// recursion for building a chain (iteration algorithm did not work
// idk why it was causing a recursion)
func compose(handler routertypes.HandlerFunc, middlewares []routertypes.Middleware) routertypes.HandlerFunc {
	if len(middlewares) == 0 {
		return handler
	}

	return func(ctx context.Context, request *types.Request) types.Response {
		return middlewares[len(middlewares)-1](
			ctx, compose(handler, middlewares[:len(middlewares)-1]), request,
		)
	}
}
