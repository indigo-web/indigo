package inbuilt

import (
	"github.com/indigo-web/indigo/v2/http/method"
	"github.com/indigo-web/indigo/v2/router/inbuilt/types"
)

/*
This file is responsible for methods shorthands - shortcuts for Route method
with already set method taken from name of the method
*/

// Get is a shortcut for registering GET-requests
func (r *Router) Get(path string, handler types.HandlerFunc, middlewares ...types.Middleware) {
	r.Route(methods.GET, path, handler, middlewares...)
}

// Head is a shortcut for registering HEAD-requests
func (r *Router) Head(path string, handler types.HandlerFunc, middlewares ...types.Middleware) {
	r.Route(methods.HEAD, path, handler, middlewares...)
}

// Post is a shortcut for registering POST-requests
func (r *Router) Post(path string, handler types.HandlerFunc, middlewares ...types.Middleware) {
	r.Route(methods.POST, path, handler, middlewares...)
}

// Put is a shortcut for registering PUT-requests
func (r *Router) Put(path string, handler types.HandlerFunc, middlewares ...types.Middleware) {
	r.Route(methods.PUT, path, handler, middlewares...)
}

// Delete is a shortcut for registering DELETE-requests
func (r *Router) Delete(path string, handler types.HandlerFunc, middlewares ...types.Middleware) {
	r.Route(methods.DELETE, path, handler, middlewares...)
}

// Connect is a shortcut for registering CONNECT-requests
func (r *Router) Connect(path string, handler types.HandlerFunc, middlewares ...types.Middleware) {
	r.Route(methods.CONNECT, path, handler, middlewares...)
}

// Options is a shortcut for registering OPTIONS-requests
func (r *Router) Options(path string, handler types.HandlerFunc, middlewares ...types.Middleware) {
	r.Route(methods.OPTIONS, path, handler, middlewares...)
}

// Trace is a shortcut for registering TRACE-requests
func (r *Router) Trace(path string, handler types.HandlerFunc, middlewares ...types.Middleware) {
	r.Route(methods.TRACE, path, handler, middlewares...)
}

// Patch is a shortcut for registering PATCH-requests
func (r *Router) Patch(path string, handler types.HandlerFunc, middlewares ...types.Middleware) {
	r.Route(methods.PATCH, path, handler, middlewares...)
}
