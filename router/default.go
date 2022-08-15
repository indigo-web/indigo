package router

import (
	"indigo/errors"
	methods "indigo/http/method"
	"indigo/http/render"
	"indigo/http/status"
	"indigo/http/url"
	"indigo/types"
	"strings"
)

type (
	Handler     func(request *types.Request) types.Response
	handlersMap map[methods.Method]Handler
	routesMap   map[url.Path]handlersMap
)

type DefaultRouter struct {
	prefix      string
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
	if path != "*" && !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	urlPath := url.Path(d.prefix + path)
	methodsMap, found := d.routes[urlPath]
	if !found {
		methodsMap = make(handlersMap)
		d.routes[urlPath] = methodsMap
	}

	methodsMap[method] = handler
}

func (d DefaultRouter) Get(path string, handler Handler) {
	d.Route(methods.GET, path, handler)
}

func (d DefaultRouter) Head(path string, handler Handler) {
	d.Route(methods.HEAD, path, handler)
}

func (d DefaultRouter) Post(path string, handler Handler) {
	d.Route(methods.POST, path, handler)
}

func (d DefaultRouter) Put(path string, handler Handler) {
	d.Route(methods.PUT, path, handler)
}

func (d DefaultRouter) Delete(path string, handler Handler) {
	d.Route(methods.DELETE, path, handler)
}

func (d DefaultRouter) Connect(path string, handler Handler) {
	d.Route(methods.CONNECT, path, handler)
}

func (d DefaultRouter) Options(path string, handler Handler) {
	d.Route(methods.OPTIONS, path, handler)
}

func (d DefaultRouter) Trace(path string, handler Handler) {
	d.Route(methods.TRACE, path, handler)
}

func (d DefaultRouter) Patch(path string, handler Handler) {
	d.Route(methods.PATCH, path, handler)
}

func (d DefaultRouter) RouteError(code status.Code, handler ErrorHandler) {
	d.errHandlers[code] = handler
}

func (d DefaultRouter) Group(prefix string) DefaultRouter {
	return DefaultRouter{
		prefix:      d.prefix + prefix,
		routes:      d.routes,
		errHandlers: d.errHandlers,
		renderer:    d.renderer,
	}
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
	var code status.Code

	switch err {
	case errors.ErrCloseConnection:
		code = status.ConnectionClose
	case errors.ErrBadRequest:
		code = status.BadRequest
	case errors.ErrTooLarge:
		code = status.RequestEntityTooLarge
	default:
		// unknown error, but for consistent behaviour we must respond with
		// something. Let it be some neutral error
		code = status.BadRequest
	}

	response := d.errHandlers[code](request)
	_ = respWriter(d.renderer.Response(request.Proto, response))
}
