package inbuilt

import "github.com/fakefloordiv/indigo/router/inbuilt/types"

/*
This file is a piece of REST across all this router
*/

type Resource struct {
	resource string
	root     *Router
}

func (r *Router) Resource(path string) Resource {
	return Resource{
		resource: path,
		root:     r,
	}
}

func (r Resource) Get(handler types.HandlerFunc, mwares ...types.Middleware) {
	r.root.Get(r.resource, handler, mwares...)
}

func (r Resource) Head(handler types.HandlerFunc, mwares ...types.Middleware) {
	r.root.Head(r.resource, handler, mwares...)
}

func (r Resource) Post(handler types.HandlerFunc, mwares ...types.Middleware) {
	r.root.Post(r.resource, handler, mwares...)
}

func (r Resource) Put(handler types.HandlerFunc, mwares ...types.Middleware) {
	r.root.Put(r.resource, handler, mwares...)
}

func (r Resource) Delete(handler types.HandlerFunc, mwares ...types.Middleware) {
	r.root.Delete(r.resource, handler, mwares...)
}

func (r Resource) Connect(handler types.HandlerFunc, mwares ...types.Middleware) {
	r.root.Connect(r.resource, handler, mwares...)
}

func (r Resource) Options(handler types.HandlerFunc, mwares ...types.Middleware) {
	r.root.Options(r.resource, handler, mwares...)
}

func (r Resource) Trace(handler types.HandlerFunc, mwares ...types.Middleware) {
	r.root.Trace(r.resource, handler, mwares...)
}

func (r Resource) Patch(handler types.HandlerFunc, mwares ...types.Middleware) {
	r.root.Patch(r.resource, handler, mwares...)
}
