package inbuilt

import (
	"github.com/indigo-web/indigo/http/method"
	"github.com/indigo-web/indigo/http/status"
	"github.com/indigo-web/indigo/router/inbuilt/types"
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
	err := r.registrar.Add(r.prefix+path, method, combine(handlerFunc, middlewares))
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
// - status.RequestHeaderFieldsTooLarge
// - status.HTTPVersionNotSupported
// - status.UnsupportedMediaType
// - status.NotImplemented
// - status.RequestTimeout
//
// You can set your own handler and override default response.
//
// WARNING: calling this method from groups will affect ALL routers, including root
func (r *Router) RouteError(handler types.Handler, codes ...status.Code) {
	for _, code := range codes {
		if code == AllErrors {
			r.errHandlers.SetUniversal(handler)
			continue
		}

		r.errHandlers.Set(code, handler)
	}
}
