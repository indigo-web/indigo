package inbuilt

import (
	"github.com/fakefloordiv/indigo/http/method"
)

/*
This file is responsible for methods predicates - shortcuts for Route method
with already set method taken from name of the method
*/

func (d Router) Get(path string, handler HandlerFunc, middlewares ...Middleware) {
	d.Route(methods.GET, path, handler, middlewares...)
}

func (d Router) Head(path string, handler HandlerFunc, middlewares ...Middleware) {
	d.Route(methods.HEAD, path, handler, middlewares...)
}

func (d Router) Post(path string, handler HandlerFunc, middlewares ...Middleware) {
	d.Route(methods.POST, path, handler, middlewares...)
}

func (d Router) Put(path string, handler HandlerFunc, middlewares ...Middleware) {
	d.Route(methods.PUT, path, handler, middlewares...)
}

func (d Router) Delete(path string, handler HandlerFunc, middlewares ...Middleware) {
	d.Route(methods.DELETE, path, handler, middlewares...)
}

func (d Router) Connect(path string, handler HandlerFunc, middlewares ...Middleware) {
	d.Route(methods.CONNECT, path, handler, middlewares...)
}

func (d Router) Options(path string, handler HandlerFunc, middlewares ...Middleware) {
	d.Route(methods.OPTIONS, path, handler, middlewares...)
}

func (d Router) Trace(path string, handler HandlerFunc, middlewares ...Middleware) {
	d.Route(methods.TRACE, path, handler, middlewares...)
}

func (d Router) Patch(path string, handler HandlerFunc, middlewares ...Middleware) {
	d.Route(methods.PATCH, path, handler, middlewares...)
}
