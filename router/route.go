package router

import (
	methods "indigo/http/method"
	"indigo/http/status"
	"indigo/http/url"
	"strings"
)

/*
This file is responsible for registering handlers
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

	urlPath := url.Path(d.prefix + path)
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
// - status.ConnectionClose
// - status.BadRequest
// - status.RequestEntityTooLarge
// - status.RequestHeaderFieldsTooLarge
// - status.RequestURITooLong
// - status.HTTPVersionNotSupported
// You can set your own handler and override default response
func (d DefaultRouter) RouteError(code status.Code, handler ErrorHandler) {
	d.errHandlers[code] = handler
}
