package router

import (
	"fmt"
	"indigo/http/render"
	"indigo/http/status"
	"indigo/http/url"
	"indigo/types"
)

type Handler func(request *types.Request) types.Response

var (
	defaultNotFound = types.WithResponse.
		WithCode(status.NotFound).
		WithBody(`<h1 align="center">404 Request Page Not Found</h1>`)
)

type DefaultRouter struct {
	routes   map[url.Path]Handler
	renderer render.Renderer
}

func NewDefaultRouter() DefaultRouter {
	return DefaultRouter{
		routes: make(map[url.Path]Handler),
		// let the first time response be rendered into the nil buffer
		renderer: render.NewRenderer(nil),
	}
}

func (d DefaultRouter) Route(path string, handler Handler) {
	d.routes[url.Path(path)] = handler
}

func (d DefaultRouter) OnRequest(request *types.Request, respWriter types.ResponseWriter) error {
	handler, found := d.routes[request.Path]
	if !found {
		return respWriter(d.renderer.Response(request.Proto, defaultNotFound))
	}

	return respWriter(d.renderer.Response(request.Proto, handler(request)))
}

func (d DefaultRouter) OnError(err error) {
	fmt.Println("[router/default] ERROR:", err)
}
