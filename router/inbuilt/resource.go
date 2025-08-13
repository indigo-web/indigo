package inbuilt

import (
	"github.com/indigo-web/indigo/http/method"
)

// Resource is just a wrapper of a group for some resource, allowing to attach
// multiple methods (and pointed-applied middlewares) to some single resource
// in a bit more convenient way than ordinary groups do. Actually, the only
// point of this object is to wrap a group to bypass an empty string into it as
// a path
type Resource struct {
	group *Router
}

// Use applies middlewares to the resource, wrapping all the already registered
// and registered in future handlers
func (r Resource) Use(middlewares ...Middleware) Resource {
	r.group.Use(middlewares...)
	return r
}

// Static adds a catcher of prefix, that automatically returns files from defined root
// directory
func (r Resource) Static(prefix, root string) Resource {
	r.group.Static(prefix, root)
	return r
}

// Route is a shortcut to group.Route, providing the extra empty path to the call
func (r Resource) Route(method method.Method, fun Handler, mwares ...Middleware) Resource {
	r.group.Route(method, "", fun, mwares...)
	return r
}

// Get registers a handler for GET-requests
func (r Resource) Get(handler Handler, mwares ...Middleware) Resource {
	r.group.Get("", handler, mwares...)
	return r
}

// Head registers a handler for HEAD-requests
func (r Resource) Head(handler Handler, mwares ...Middleware) Resource {
	r.group.Head("", handler, mwares...)
	return r
}

// Post registers a handler for POST-requests
func (r Resource) Post(handler Handler, mwares ...Middleware) Resource {
	r.group.Post("", handler, mwares...)
	return r
}

// Put registers a handler for PUT-requests
func (r Resource) Put(handler Handler, mwares ...Middleware) Resource {
	r.group.Put("", handler, mwares...)
	return r
}

// Delete registers a handler for DELETE-requests
func (r Resource) Delete(handler Handler, mwares ...Middleware) Resource {
	r.group.Delete("", handler, mwares...)
	return r
}

// Connect registers a handler for CONNECT-requests
func (r Resource) Connect(handler Handler, mwares ...Middleware) Resource {
	r.group.Connect("", handler, mwares...)
	return r
}

// Options registers a handler for OPTIONS-requests
func (r Resource) Options(handler Handler, mwares ...Middleware) Resource {
	r.group.Options("", handler, mwares...)
	return r
}

// Trace registers a handler for TRACE-requests
func (r Resource) Trace(handler Handler, mwares ...Middleware) Resource {
	r.group.Trace("", handler, mwares...)
	return r
}

// Patch registers a handler for PATCH-requests
func (r Resource) Patch(handler Handler, mwares ...Middleware) Resource {
	r.group.Patch("", handler, mwares...)
	return r
}

// Mkcol registers a handler for MKCOL-requests
func (r Resource) Mkcol(handler Handler, mwares ...Middleware) Resource {
	r.group.Mkcol("", handler, mwares...)
	return r
}

// Move registers a handler for MOVE-requests
func (r Resource) Move(handler Handler, mwares ...Middleware) Resource {
	r.group.Move("", handler, mwares...)
	return r
}

// Copy registers a handler for COPY-requests
func (r Resource) Copy(handler Handler, mwares ...Middleware) Resource {
	r.group.Copy("", handler, mwares...)
	return r
}

// Lock registers a handler for LOCK-requests
func (r Resource) Lock(handler Handler, mwares ...Middleware) Resource {
	r.group.Lock("", handler, mwares...)
	return r
}

// Unlock registers a handler for UNLOCK-requests
func (r Resource) Unlock(handler Handler, mwares ...Middleware) Resource {
	r.group.Unlock("", handler, mwares...)
	return r
}

// Propfind registers a handler for PROPFIND-requests
func (r Resource) Propfind(handler Handler, mwares ...Middleware) Resource {
	r.group.Propfind("", handler, mwares...)
	return r
}

// Proppatch registers a handler for PROPPATCH-requests
func (r Resource) Proppatch(handler Handler, mwares ...Middleware) Resource {
	r.group.Proppatch("", handler, mwares...)
	return r
}
