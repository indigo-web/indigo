package inbuilt

import (
	"fmt"
	"path"
)

// Catcher is used to catch requests, if no other handlers are available. This
// is used, for example, for static files distribution
type Catcher struct {
	Prefix  string
	Handler Handler
}

// Catch registers a catcher. A catcher is a handler, that is being called if requested path
// is not found, and it starts with a defined prefix
func (r *Router) Catch(prefix string, handler Handler, middlewares ...Middleware) *Router {
	prefix = path.Join(r.prefix, prefix)

	for _, catcher := range r.catchers {
		if catcher.Prefix == prefix {
			panic(fmt.Errorf("catcher already registered: %s", prefix))
		}
	}

	r.catchers = append(r.catchers, Catcher{
		Prefix:  prefix,
		Handler: compose(handler, middlewares),
	})

	return r
}

func (r *Router) applyCatchersMiddlewares() {
	for i := range r.catchers {
		r.catchers[i].Handler = compose(r.catchers[i].Handler, r.middlewares)
	}
}
