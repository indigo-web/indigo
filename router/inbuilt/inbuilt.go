package inbuilt

import (
	"context"
	methods "github.com/fakefloordiv/indigo/http/method"
	"github.com/fakefloordiv/indigo/types"
)

type (
	HandlerFunc func(context.Context, *types.Request) types.Response
	handlersMap map[methods.Method]*handlerObject
	routesMap   map[string]handlersMap

	handlerObject struct {
		fun         HandlerFunc
		middlewares []Middleware
	}

	ErrorHandler func(context.Context, *types.Request) types.Response
	errHandlers  map[error]ErrorHandler
)

// Router is a reference implementation of router for indigo
// It supports:
// 1) Endpoint groups
// 2) Middlewares
// 3) Error handlers
// 4) Encoding/decoding incoming content
// 5) Routing by path and method. If path not found, 404 Not Found is returned.
//    If path is found, but no method attached, 413 Method Not Allowed is returned.
type Router struct {
	root   *Router
	groups []Router

	prefix      string
	middlewares []Middleware

	routes         routesMap
	errHandlers    errHandlers
	allowedMethods map[string]string
}

// NewRouter constructs a new instance of inbuilt router. Error handlers
// by default are applied, renderer with a nil (as initial value) buffer constructed
func NewRouter() *Router {
	r := &Router{
		routes:         make(routesMap),
		errHandlers:    newErrHandlers(),
		allowedMethods: make(map[string]string),
	}

	r.root = r

	return r
}
