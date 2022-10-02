package inbuilt

import (
	"context"
	"github.com/fakefloordiv/indigo/router/inbuilt/obtainer"
	routertypes "github.com/fakefloordiv/indigo/router/inbuilt/types"
	"github.com/fakefloordiv/indigo/types"
)

type (
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
	middlewares []routertypes.Middleware

	obtainer obtainer.Obtainer

	routes      routertypes.RoutesMap
	errHandlers errHandlers

	traceBuff []byte
}

// NewRouter constructs a new instance of inbuilt router. Error handlers
// by default are applied, renderer with a nil (as initial value) buffer constructed
func NewRouter() *Router {
	r := &Router{
		routes:      make(routertypes.RoutesMap),
		errHandlers: newErrorHandlers(),
	}

	r.root = r

	return r
}
