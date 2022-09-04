package inbuilt

import (
	"strings"

	methods "github.com/fakefloordiv/indigo/http/method"
)

/*
This file is responsible for registering both ordinary and error handlers
*/

// Route is a base method for registering handlers
func (d *DefaultRouter) Route(
	method methods.Method, path string, handlerFunc HandlerFunc,
	middlewares ...Middleware,
) {
	if path != "*" && !strings.HasPrefix(path, "/") && d.prefix == "" {
		// applying prefix slash only if we are not in group
		path = "/" + path
	}

	urlPath := d.prefix + path
	methodsMap, found := d.routes[urlPath]
	if !found {
		methodsMap = make(handlersMap)
		d.routes[urlPath] = methodsMap
	}

	handlerStruct := &handlerObject{
		fun:         handlerFunc,
		middlewares: append(middlewares, d.middlewares...),
	}

	methodsMap[method] = handlerStruct
}

// RouteError adds an error handler. You can handle next errors:
// - http.ErrCloseConnection
// - http.ErrBadRequest
// - http.ErrMethodNotImplemented
// - http.ErrTooLarge
// - http.ErrHeaderFieldsTooLarge
// - http.ErrURITooLong
// - http.ErrUnsupportedProtocol
// - http.ErrUnsupportedEncoding
//
// You can set your own handler and override default response
func (d DefaultRouter) RouteError(err error, handler ErrorHandler) {
	d.root.errHandlers[err] = handler
}
