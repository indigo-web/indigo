package inbuilt

import (
	"strings"

	methods "github.com/fakefloordiv/indigo/http/method"
)

/*
This file is responsible for registering both ordinary and error handlers
*/

// Route is a base method for registering handlers
func (r *Router) Route(
	method methods.Method, path string, handlerFunc HandlerFunc,
	middlewares ...Middleware,
) {
	if path != "*" && !strings.HasPrefix(path, "/") && r.prefix == "" {
		// applying prefix slash only if we are not in group
		path = "/" + path
	}

	urlPath := r.prefix + path
	methodsMap, found := r.routes[urlPath]
	if !found {
		methodsMap = make(handlersMap)
		r.routes[urlPath] = methodsMap
	}

	handlerStruct := &handlerObject{
		fun:         handlerFunc,
		middlewares: append(middlewares, r.middlewares...),
	}

	methodsMap[method] = handlerStruct
}

// RouteError adds an error handler. You can handle next errors:
// - http.ErrBadRequest
// - http.ErrNotFound
// - http.ErrMethodNotAllowed
// - http.ErrTooLarge
// - http.ErrCloseConnection
// - http.ErrURITooLong
// - http.ErrHeaderFieldsTooLarge
// - http.ErrTooManyHeaders
// - http.ErrUnsupportedProtocol
// - http.ErrUnsupportedEncoding
// - http.ErrMethodNotImplemented
// - http.ErrConnectionTimeout
//
// You can set your own handler and override default response
func (r Router) RouteError(err error, handler ErrorHandler) {
	r.root.errHandlers[err] = handler
}
