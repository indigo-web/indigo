package inbuilt

import (
	"github.com/indigo-web/indigo/router"
	"github.com/indigo-web/indigo/router/inbuilt/obtainer"
	"github.com/indigo-web/indigo/router/inbuilt/types"
)

var _ router.Router = &Router{}

// Router is a built-in implementation of router.Router interface that provides
// some basic router features like middlewares, groups, dynamic routing, error
// handlers, and some implicit things like calling GET-handlers for HEAD-requests,
// or rendering TRACE-responses automatically in case no handler is registered
type Router struct {
	root   *Router
	groups []Router

	prefix      string
	middlewares []types.Middleware

	obtainer obtainer.Obtainer

	routes      types.RoutesMap
	errHandlers types.ErrHandlers

	traceBuff []byte
}

// NewRouter constructs a new instance of inbuilt router
func NewRouter() *Router {
	r := &Router{
		routes:      make(types.RoutesMap),
		errHandlers: newErrorHandlers(),
	}

	r.root = r

	return r
}
