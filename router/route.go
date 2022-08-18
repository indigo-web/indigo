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

// Route is a base method for registering handlers. At the moment it composes
// all the handler-specific and global middlewares right now, but in future
// planning to change api, and contain Handler struct that contains handler
// middlewares that will be applied only after server started up
func (d *DefaultRouter) Route(
	method methods.Method, path string, handler HandlerFunc,
	middlewares ...Middleware,
) {
	if path != "*" && !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	urlPath := url.Path(d.prefix + path)
	methodsMap, found := d.routes[urlPath]
	if !found {
		methodsMap = make(handlersMap)
		d.routes[urlPath] = methodsMap
	}

	methodsMap[method] = compose(handler, append(middlewares, d.middlewares...))
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
