package inbuilt

import (
	"github.com/indigo-web/indigo/http/method"
	"github.com/indigo-web/indigo/http/status"
)

// AllErrors is used to be passed into Router.RouteError, indicating by that,
// that the handler must handle ALL errors (if concrete error's handler won't
// override it)
const AllErrors = status.Code(0)

// Route is a base method for registering handlers
func (r *Router) Route(
	method method.Method, path string, handlerFunc Handler, middlewares ...Middleware,
) *Router {
	err := r.registrar.Add(r.prefix+path, method, compose(handlerFunc, middlewares))
	if err != nil {
		panic(err)
	}

	return r
}

// RouteError adds an error handler for a corresponding HTTP error code.
//
// The following error codes may be registered:
//   - AllErrors (called only if no other error handlers found)
//   - status.BadRequest
//   - status.NotFound
//   - status.MethodNotAllowed
//   - status.RequestEntityTooLarge
//   - status.CloseConnection
//   - status.RequestURITooLong
//   - status.HeaderFieldsTooLarge
//   - status.HTTPVersionNotSupported
//   - status.UnsupportedMediaType
//   - status.NotImplemented
//   - status.RequestTimeout
//
// Note: if handler returned one of error codes above, error handler WON'T be called.
// Also, global middlewares, applied to the root router, will also be used for error handlers.
// However, global middlewares defined on groups won't be used.
//
// WARNING: calling this method from groups will affect ALL routers, including root
func (r *Router) RouteError(handler Handler, codes ...status.Code) *Router {
	if len(codes) == 0 {
		codes = append(codes, AllErrors)
	}

	for _, code := range codes {
		r.errHandlers[code] = handler
	}

	return r
}
