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

// Catch registers a catcher. A catcher is a handler, that is being called if requested path
// is not found, and it starts with a defined prefix
func (r Resource) Catch(prefix string, handler types.Handler, mwares ...types.Middleware) Resource {
	r.group.Catch(prefix, handler, mwares...)
	return r
}

// Route is a shortcut to group.Route, providing the extra empty path to the call
func (r Resource) Route(method method.Method, fun types.Handler, mwares ...types.Middleware) Resource {
	r.group.Route(method, "", fun, mwares...)
	return r
}

// Get registers a handler for GET-requests
func (r Resource) Get(handler types.Handler, mwares ...types.Middleware) Resource {
	r.group.Get("", handler, mwares...)
	return r
}

// Head registers a handler for HEAD-requests
func (r Resource) Head(handler types.Handler, mwares ...types.Middleware) Resource {
	r.group.Head("", handler, mwares...)
	return r
}

// Post registers a handler for POST-requests
func (r Resource) Post(handler types.Handler, mwares ...types.Middleware) Resource {
	r.group.Post("", handler, mwares...)
	return r
}

// Put registers a handler for PUT-requests
func (r Resource) Put(handler types.Handler, mwares ...types.Middleware) Resource {
	r.group.Put("", handler, mwares...)
	return r
}

// Delete registers a handler for DELETE-requests
func (r Resource) Delete(handler types.Handler, mwares ...types.Middleware) Resource {
	r.group.Delete("", handler, mwares...)
	return r
}

// Connect registers a handler for CONNECT-requests
func (r Resource) Connect(handler types.Handler, mwares ...types.Middleware) Resource {
	r.group.Connect("", handler, mwares...)
	return r
}

// Options registers a handler for OPTIONS-requests
func (r Resource) Options(handler types.Handler, mwares ...types.Middleware) Resource {
	r.group.Options("", handler, mwares...)
	return r
}

// Trace registers a handler for TRACE-requests
func (r Resource) Trace(handler types.Handler, mwares ...types.Middleware) Resource {
	r.group.Trace("", handler, mwares...)
	return r
}

// Patch registers a handler for PATCH-requests
func (r Resource) Patch(handler types.Handler, mwares ...types.Middleware) Resource {
	r.group.Patch("", handler, mwares...)
	return r
}
