package inbuilt

import (
	"github.com/fakefloordiv/indigo/http/method"
	"github.com/fakefloordiv/indigo/router/inbuilt/types"
)

/*
This file is responsible for methods shorthands - shortcuts for Route method
with already set method taken from name of the method
*/

func (r *Router) Get(path string, handler types.HandlerFunc, middlewares ...types.Middleware) {
	r.Route(methods.GET, path, handler, middlewares...)
}

func (r *Router) Head(path string, handler types.HandlerFunc, middlewares ...types.Middleware) {
	r.Route(methods.HEAD, path, handler, middlewares...)
}

func (r *Router) Post(path string, handler types.HandlerFunc, middlewares ...types.Middleware) {
	r.Route(methods.POST, path, handler, middlewares...)
}

func (r *Router) Put(path string, handler types.HandlerFunc, middlewares ...types.Middleware) {
	r.Route(methods.PUT, path, handler, middlewares...)
}

func (r *Router) Delete(path string, handler types.HandlerFunc, middlewares ...types.Middleware) {
	r.Route(methods.DELETE, path, handler, middlewares...)
}

func (r *Router) Connect(path string, handler types.HandlerFunc, middlewares ...types.Middleware) {
	r.Route(methods.CONNECT, path, handler, middlewares...)
}

func (r *Router) Options(path string, handler types.HandlerFunc, middlewares ...types.Middleware) {
	r.Route(methods.OPTIONS, path, handler, middlewares...)
}

func (r *Router) Trace(path string, handler types.HandlerFunc, middlewares ...types.Middleware) {
	r.Route(methods.TRACE, path, handler, middlewares...)
}

func (r *Router) Patch(path string, handler types.HandlerFunc, middlewares ...types.Middleware) {
	r.Route(methods.PATCH, path, handler, middlewares...)
}
