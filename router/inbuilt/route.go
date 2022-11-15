package inbuilt

import (
	methods "github.com/fakefloordiv/indigo/http/method"
	routertypes "github.com/fakefloordiv/indigo/router/inbuilt/types"
)

/*
This file is responsible for registering both ordinary and error handlers
*/

// Route is a base method for registering handlers
func (r *Router) Route(
	method methods.Method, path string, handlerFunc routertypes.HandlerFunc,
	middlewares ...routertypes.Middleware,
) {
	urlPath := r.prefix + path
	methodsMap, found := r.routes[urlPath]
	if !found {
		methodsMap = make(routertypes.MethodsMap)
		r.routes[urlPath] = methodsMap
	}

	handlerStruct := &routertypes.HandlerObject{
		Fun:         handlerFunc,
		Middlewares: append(middlewares, r.middlewares...),
	}

	methodsMap[method] = handlerStruct
}

// RouteError adds an error handler. You can handle next errors:
// - status.ErrBadRequest
// - status.ErrNotFound
// - status.ErrMethodNotAllowed
// - status.ErrTooLarge
// - status.ErrCloseConnection
// - status.ErrURITooLong
// - status.ErrHeaderFieldsTooLarge
// - status.ErrTooManyHeaders
// - status.ErrUnsupportedProtocol
// - status.ErrUnsupportedEncoding
// - status.ErrMethodNotImplemented
// - status.ErrConnectionTimeout
//
// You can set your own handler and override default response
func (r Router) RouteError(err error, handler routertypes.HandlerFunc) {
	r.root.errHandlers[err] = handler
}
