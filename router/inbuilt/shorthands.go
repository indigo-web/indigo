package inbuilt

import (
	"github.com/fakefloordiv/indigo/http/method"
)

/*
This file is responsible for methods predicates - shortcuts for Route method
with already set method taken from name of the method
*/

func (r Router) Get(path string, handler HandlerFunc, middlewares ...Middleware) {
	r.Route(methods.GET, path, handler, middlewares...)
}

func (r Router) Head(path string, handler HandlerFunc, middlewares ...Middleware) {
	r.Route(methods.HEAD, path, handler, middlewares...)
}

func (r Router) Post(path string, handler HandlerFunc, middlewares ...Middleware) {
	r.Route(methods.POST, path, handler, middlewares...)
}

func (r Router) Put(path string, handler HandlerFunc, middlewares ...Middleware) {
	r.Route(methods.PUT, path, handler, middlewares...)
}

func (r Router) Delete(path string, handler HandlerFunc, middlewares ...Middleware) {
	r.Route(methods.DELETE, path, handler, middlewares...)
}

func (r Router) Connect(path string, handler HandlerFunc, middlewares ...Middleware) {
	r.Route(methods.CONNECT, path, handler, middlewares...)
}

func (r Router) Options(path string, handler HandlerFunc, middlewares ...Middleware) {
	r.Route(methods.OPTIONS, path, handler, middlewares...)
}

func (r Router) Trace(path string, handler HandlerFunc, middlewares ...Middleware) {
	r.Route(methods.TRACE, path, handler, middlewares...)
}

func (r Router) Patch(path string, handler HandlerFunc, middlewares ...Middleware) {
	r.Route(methods.PATCH, path, handler, middlewares...)
}
