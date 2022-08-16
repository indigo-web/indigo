package router

import (
	methods "indigo/http/method"
	"indigo/http/status"
	"indigo/http/url"
	"strings"
)

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

func (d DefaultRouter) RouteError(code status.Code, handler ErrorHandler) {
	d.errHandlers[code] = handler
}
