package inbuilt

import (
	"fmt"
	"github.com/indigo-web/iter"
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
	prefixes := iter.Map[Catcher, bool](func(el Catcher) bool {
		return el.Prefix == prefix
	}, iter.Slice(r.catchers))

	if iter.Reduce[bool](func(prev, curr bool) bool {
		return prev || curr
	}, prefixes) {
		panic(fmt.Errorf("catcher already exists: %s", prefix))
	}

	r.catchers = append(r.catchers, Catcher{
		Prefix:  prefix,
		Handler: compose(handler, middlewares),
	})

	return r
}
