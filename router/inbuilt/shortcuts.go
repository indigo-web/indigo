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

// Mkcol is a shortcut for registering MKCOL-requests.
func (r *Router) Mkcol(path string, handler Handler, middlewares ...Middleware) *Router {
	r.Route(method.MKCOL, path, handler, middlewares...)
	return r
}

// Move is a shortcut for registering MOVE-requests.
func (r *Router) Move(path string, handler Handler, middlewares ...Middleware) *Router {
	r.Route(method.MOVE, path, handler, middlewares...)
	return r
}

// Copy is a shortcut for registering COPY-requests.
func (r *Router) Copy(path string, handler Handler, middlewares ...Middleware) *Router {
	r.Route(method.COPY, path, handler, middlewares...)
	return r
}

// Lock is a shortcut for registering LOCK-requests.
func (r *Router) Lock(path string, handler Handler, middlewares ...Middleware) *Router {
	r.Route(method.LOCK, path, handler, middlewares...)
	return r
}

// Unlock is a shortcut for registering UNLOCK-requests.
func (r *Router) Unlock(path string, handler Handler, middlewares ...Middleware) *Router {
	r.Route(method.UNLOCK, path, handler, middlewares...)
	return r
}

// Propfind is a shortcut for registering PROPFIND-requests.
func (r *Router) Propfind(path string, handler Handler, middlewares ...Middleware) *Router {
	r.Route(method.PROPFIND, path, handler, middlewares...)
	return r
}

// Proppatch is a shortcut for registering PROPPATCH-requests.
func (r *Router) Proppatch(path string, handler Handler, middlewares ...Middleware) *Router {
	r.Route(method.PROPPATCH, path, handler, middlewares...)
	return r
}

// File is a shortcut handler for single file endpoints.
func File(filename string) Handler {
	return func(request *http.Request) *http.Response {
		return http.File(request, filename)
	}
}
