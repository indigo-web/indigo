package inbuilt

import (
	"github.com/indigo-web/indigo/http"
	"github.com/indigo-web/indigo/http/method"
)

// Get is a shortcut for registering GET-requests.
func (r *Router) Get(path string, handler Handler, middlewares ...Middleware) *Router {
	r.Route(method.GET, path, handler, middlewares...)
	return r
}

// Head is a shortcut for registering HEAD-requests.
func (r *Router) Head(path string, handler Handler, middlewares ...Middleware) *Router {
	r.Route(method.HEAD, path, handler, middlewares...)
	return r
}

// Post is a shortcut for registering POST-requests.
func (r *Router) Post(path string, handler Handler, middlewares ...Middleware) *Router {
	r.Route(method.POST, path, handler, middlewares...)
	return r
}

// Put is a shortcut for registering PUT-requests.
func (r *Router) Put(path string, handler Handler, middlewares ...Middleware) *Router {
	r.Route(method.PUT, path, handler, middlewares...)
	return r
}

// Delete is a shortcut for registering DELETE-requests.
func (r *Router) Delete(path string, handler Handler, middlewares ...Middleware) *Router {
	r.Route(method.DELETE, path, handler, middlewares...)
	return r
}

// Connect is a shortcut for registering CONNECT-requests.
func (r *Router) Connect(path string, handler Handler, middlewares ...Middleware) *Router {
	r.Route(method.CONNECT, path, handler, middlewares...)
	return r
}

// Options is a shortcut for registering OPTIONS-requests.
func (r *Router) Options(path string, handler Handler, middlewares ...Middleware) *Router {
	r.Route(method.OPTIONS, path, handler, middlewares...)
	return r
}

// Trace is a shortcut for registering TRACE-requests.
func (r *Router) Trace(path string, handler Handler, middlewares ...Middleware) *Router {
	r.Route(method.TRACE, path, handler, middlewares...)
	return r
}

// Patch is a shortcut for registering PATCH-requests.
func (r *Router) Patch(path string, handler Handler, middlewares ...Middleware) *Router {
	r.Route(method.PATCH, path, handler, middlewares...)
	return r
}

// File is a shortcut for responding with a single file without having to write own
// closure or enabling static content distribution.
func File(filename string) Handler {
	return func(request *http.Request) *http.Response {
		return http.File(request, filename)
	}
}
