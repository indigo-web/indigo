package inbuilt

import (
	methods "github.com/fakefloordiv/indigo/http/method"
	"github.com/fakefloordiv/indigo/types"
)

type (
	HandlerFunc func(*types.Request) types.Response
	handlersMap map[methods.Method]*handlerObject
	routesMap   map[string]handlersMap

	handlerObject struct {
		fun         HandlerFunc
		middlewares []Middleware
	}

	ErrorHandler func(request *types.Request) types.Response
	errHandlers  map[error]ErrorHandler
)

// DefaultRouter is a reference implementation of router for indigo
// It supports:
// 1) Endpoint groups
// 2) Middlewares
// 3) Error handlers
// 4) Encoding/decoding incoming content
// 5) Routing by path and method. If path not found, 404 Not Found is returned.
//    If path is found, but no method attached, 413 Method Not Allowed is returned.
type DefaultRouter struct {
	root   *DefaultRouter
	groups []DefaultRouter

	prefix      string
	middlewares []Middleware

	routes      routesMap
	errHandlers errHandlers
}

// NewRouter constructs a new instance of inbuilt router. Error handlers
// by default are applied, renderer with a nil (as initial value) buffer constructed
func NewRouter() *DefaultRouter {
	r := &DefaultRouter{
		routes:      make(routesMap),
		errHandlers: newErrHandlers(),
	}

	r.root = r

	return r
}
