package router

import "indigo/types"

type Middleware func(next HandlerFunc, request *types.Request) types.Response

func (d *DefaultRouter) Use(middleware Middleware) {
	d.middlewares = append(d.middlewares, middleware)
}

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
