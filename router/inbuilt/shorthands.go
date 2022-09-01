package inbuilt

import (
	"indigo/http/method"
)

/*
This file is responsible for methods predicates - shortcuts for Route method
with already set method taken from name of the method
*/

func (d DefaultRouter) Get(path string, handler HandlerFunc, middlewares ...Middleware) {
	d.Route(methods.GET, path, handler, middlewares...)
}

func (d DefaultRouter) Head(path string, handler HandlerFunc, middlewares ...Middleware) {
	d.Route(methods.HEAD, path, handler, middlewares...)
}

func (d DefaultRouter) Post(path string, handler HandlerFunc, middlewares ...Middleware) {
	d.Route(methods.POST, path, handler, middlewares...)
}

func (d DefaultRouter) Put(path string, handler HandlerFunc, middlewares ...Middleware) {
	d.Route(methods.PUT, path, handler, middlewares...)
}

func (d DefaultRouter) Delete(path string, handler HandlerFunc, middlewares ...Middleware) {
	d.Route(methods.DELETE, path, handler, middlewares...)
}

func (d DefaultRouter) Connect(path string, handler HandlerFunc, middlewares ...Middleware) {
	d.Route(methods.CONNECT, path, handler, middlewares...)
}

func (d DefaultRouter) Options(path string, handler HandlerFunc, middlewares ...Middleware) {
	d.Route(methods.OPTIONS, path, handler, middlewares...)
}

func (d DefaultRouter) Trace(path string, handler HandlerFunc, middlewares ...Middleware) {
	d.Route(methods.TRACE, path, handler, middlewares...)
}

func (d DefaultRouter) Patch(path string, handler HandlerFunc, middlewares ...Middleware) {
	d.Route(methods.PATCH, path, handler, middlewares...)
}
