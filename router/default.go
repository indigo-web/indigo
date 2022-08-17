package router

import (
	"indigo/http/encodings"
	methods "indigo/http/method"
	"indigo/http/render"
	"indigo/http/url"
	"indigo/types"
)

type (
	HandlerFunc func(*types.Request) types.Response
	handlersMap map[methods.Method]HandlerFunc
	routesMap   map[url.Path]handlersMap
)

type DefaultRouter struct {
	prefix      string
	middlewares []Middleware

	routes      routesMap
	errHandlers errHandlers

	renderer *render.Renderer
	codings  *encodings.ContentEncodings
}

func NewDefaultRouter() DefaultRouter {
	contentEncodings := encodings.NewContentEncodings()

	return DefaultRouter{
		routes:      make(routesMap),
		errHandlers: newErrHandlers(),
		// let the first time response be rendered into the nil buffer
		renderer: render.NewRenderer(nil),
		codings:  &contentEncodings,
	}
}
