package inbuilt

import "path"

// Catcher is used to catch requests, if no other handlers are available. This
// is used, for example, for static files distribution
type Catcher struct {
	Prefix  string
	Handler Handler
}

// Catch registers a catcher. A catcher is a handler, that is being called if requested path
// is not found, and it starts with a defined prefix
func (r *Router) Catch(prefix string, handler Handler, middlewares ...Middleware) *Router {
	r.catchers = append(r.catchers, Catcher{
		Prefix:  path.Join(r.prefix, prefix),
		Handler: compose(handler, middlewares),
	})

	return r
}
