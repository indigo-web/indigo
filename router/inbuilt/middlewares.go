package inbuilt

import (
	"github.com/indigo-web/indigo/http"
)

// Middleware works like a chain of nested calls, next may be even directly
// handler. But if we are not a closing middleware, we will call next
// middleware that is simply a partial middleware with already provided next
type Middleware func(next Handler, request *http.Request) *http.Response

// Use adds middlewares into the global list of a group's middlewares. But they will
// be applied only after the server will be started
func (r *Router) Use(middlewares ...Middleware) *Router {
	r.middlewares = append(r.middlewares, middlewares...)

	return r
}

func (r *Router) applyMiddlewares() {
	r.registrar.Apply(func(handler Handler) Handler {
		return compose(handler, r.middlewares)
	})
}

// compose produces an array of middlewares into the chain, represented by types.Handler
func compose(handler Handler, middlewares []Middleware) Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		handler = func(handler Handler, middleware Middleware) Handler {
			return func(request *http.Request) *http.Response {
				return middleware(handler, request)
			}
		}(handler, middlewares[i])
	}

	return handler
}
