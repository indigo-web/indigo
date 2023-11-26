package inbuilt

import (
	"github.com/indigo-web/indigo/http/method"
	"github.com/indigo-web/indigo/http/status"
	"github.com/indigo-web/indigo/router/inbuilt/types"
	"path"
)

// AllErrors is used to be passed into Router.RouteError, indicating by that,
// that the handler must handle ALL errors (if concrete error's handler won't
// override it)
const AllErrors = status.Code(0)

// Route is a base method for registering handlers
func (r *Router) Route(
	method method.Method, path string, handlerFunc types.Handler,
	middlewares ...types.Middleware,
) *Router {
	err := r.registrar.Add(r.prefix+path, method, compose(handlerFunc, middlewares))
	if err != nil {
		panic(err)
	}

	return r
}

// RouteError adds an error handler for a corresponding HTTP error code.
//
// Note: error codes are only 4xx and 5xx. Registering for other codes will result
// in panicking.
//
// The following error codes may be handled:
// - AllErrors
// - status.BadRequest
// - status.NotFound
// - status.MethodNotAllowed
// - status.RequestEntityTooLarge
// - status.CloseConnection
// - status.RequestURITooLong
// - status.HeaderFieldsTooLarge
// - status.HTTPVersionNotSupported
// - status.UnsupportedMediaType
// - status.NotImplemented
// - status.RequestTimeout
//
// You can set your own handler and override default response.
//
// WARNING: calling this method from groups will affect ALL routers, including root
func (r *Router) RouteError(handler types.Handler, codes ...status.Code) *Router {
	for _, code := range codes {
		if code == AllErrors {
			r.errHandlers.SetUniversal(handler)
			continue
		}

		r.errHandlers.Set(code, handler)
	}

	return r
}

// Catch registers a catcher. A catcher is a handler, that is being called if requested path
// is not found, and it starts with a defined prefix
func (r *Router) Catch(prefix string, handler types.Handler, middlewares ...types.Middleware) *Router {
	r.catchers = append(r.catchers, types.Catcher{
		Prefix:  path.Join(r.prefix, prefix),
		Handler: compose(handler, middlewares),
	})

	return r
}
