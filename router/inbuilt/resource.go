package inbuilt

import (
	"github.com/indigo-web/indigo/http/method"
	"github.com/indigo-web/indigo/router/inbuilt/types"
)

// Resource is just a wrapper of a group for some resource, allowing to attach
// multiple methods (and pointed-applied middlewares) to some single resource
// in a bit more convenient way than ordinary groups do. Actually, the only
// point of this object is to wrap a group to bypass an empty string into it as
// a path
type Resource struct {
	group *Router
}

// Resource returns a new Resource object for a provided resource path
func (r *Router) Resource(path string) Resource {
	return Resource{
		group: r.Group(path),
	}
}

// Use applies middlewares to the resource, wrapping all the already registered
// and registered in future handlers
func (r Resource) Use(middlewares ...types.Middleware) Resource {
	r.group.Use(middlewares...)
	return r
}

// Route is a shortcut to group.Route, providing the extra empty path to the call
func (r Resource) Route(method method.Method, fun types.HandlerFunc, mwares ...types.Middleware) Resource {
	r.group.Route(method, "", fun, mwares...)
	return r
}

// Get is a shortcut for registering GET-requests
func (r Resource) Get(handler types.HandlerFunc, mwares ...types.Middleware) Resource {
	r.group.Get("", handler, mwares...)
	return r
}

// Head is a shortcut for registering HEAD-requests
func (r Resource) Head(handler types.HandlerFunc, mwares ...types.Middleware) Resource {
	r.group.Head("", handler, mwares...)
	return r
}

// Post is a shortcut for registering POST-requests
func (r Resource) Post(handler types.HandlerFunc, mwares ...types.Middleware) Resource {
	r.group.Post("", handler, mwares...)
	return r
}

// Put is a shortcut for registering PUT-requests
func (r Resource) Put(handler types.HandlerFunc, mwares ...types.Middleware) Resource {
	r.group.Put("", handler, mwares...)
	return r
}

// Delete is a shortcut for registering DELETE-requests
func (r Resource) Delete(handler types.HandlerFunc, mwares ...types.Middleware) Resource {
	r.group.Delete("", handler, mwares...)
	return r
}

// Connect is a shortcut for registering CONNECT-requests
func (r Resource) Connect(handler types.HandlerFunc, mwares ...types.Middleware) Resource {
	r.group.Connect("", handler, mwares...)
	return r
}

// Options is a shortcut for registering OPTIONS-requests
func (r Resource) Options(handler types.HandlerFunc, mwares ...types.Middleware) Resource {
	r.group.Options("", handler, mwares...)
	return r
}

// Trace is a shortcut for registering TRACE-requests
func (r Resource) Trace(handler types.HandlerFunc, mwares ...types.Middleware) Resource {
	r.group.Trace("", handler, mwares...)
	return r
}

// Patch is a shortcut for registering PATCH-requests
func (r Resource) Patch(handler types.HandlerFunc, mwares ...types.Middleware) Resource {
	r.group.Patch("", handler, mwares...)
	return r
}
