package inbuilt

import (
	"github.com/indigo-web/indigo/http"
	"github.com/indigo-web/indigo/router/inbuilt/types"
)

/*
This file is responsible for middlewares
*/

// Use adds middlewares into the global list of a group's middlewares. But they will
// be applied only after the server will be started
func (r *Router) Use(middlewares ...types.Middleware) *Router {
	r.middlewares = append(r.middlewares, middlewares...)

	return r
}

func (r *Router) applyMiddlewares() {
	r.registrar.Apply(func(handler types.Handler) types.Handler {
		return compose(handler, r.middlewares)
	})
}

// compose produces an array of middlewares into the chain, represented by types.Handler
func compose(handler types.Handler, middlewares []types.Middleware) types.Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		handler = func(handler types.Handler, middleware types.Middleware) types.Handler {
			return func(request *http.Request) *http.Response {
				return middleware(handler, request)
			}
		}(handler, middlewares[i])
	}

	return handler
}
