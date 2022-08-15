package router

import (
	"indigo/errors"
	methods "indigo/http/method"
	"indigo/http/render"
	"indigo/http/status"
	"indigo/http/url"
	"indigo/types"
)

type (
	Handler     func(request *types.Request) types.Response
	handlersMap map[methods.Method]Handler
	routesMap   map[url.Path]handlersMap
)

type DefaultRouter struct {
	routes      routesMap
	errHandlers errHandlers
	renderer    render.Renderer
}

func NewDefaultRouter() DefaultRouter {
	return DefaultRouter{
		routes:      make(routesMap),
		errHandlers: newErrHandlers(),
		// let the first time response be rendered into the nil buffer
		renderer: render.NewRenderer(nil),
	}
}

func (d DefaultRouter) Route(method methods.Method, path string, handler Handler) {
	urlPath := url.Path(path)
	methodsMap, found := d.routes[urlPath]
	if !found {
		d.routes[urlPath] = handlersMap{}
	}

	methodsMap[method] = handler
}

func (d DefaultRouter) RouteError(code status.Code, handler ErrorHandler) {
	d.errHandlers[code] = handler
}

func (d DefaultRouter) OnRequest(request *types.Request, respWriter types.ResponseWriter) error {
	urlMethods, found := d.routes[request.Path]
	if !found {
		return respWriter(d.renderer.Response(request.Proto, defaultNotFound))
	}

	handler, found := urlMethods[request.Method]
	if !found {
		return respWriter(d.renderer.Response(request.Proto, defaultMethodNotAllowed))
	}

	return respWriter(d.renderer.Response(request.Proto, handler(request)))
}

func (d DefaultRouter) OnError(request *types.Request, respWriter types.ResponseWriter, err error) {
	// currently this callback may be only called with these 2 errors
	switch err {
	case errors.ErrCloseConnection:
		d.errHandlers[status.ConnectionClose](request)
	case errors.ErrBadRequest:
		response := d.errHandlers[status.BadRequest](request)
		// we don't care whether any error occurred here because in case of calling this
		// callback connection will be anyway closed
		_ = respWriter(d.renderer.Response(request.Proto, response))
	}
}
