package inbuilt

import (
	"github.com/indigo-web/indigo/router/inbuilt/mutator"
	"path"
)

// Alias makes an implicitly redirects to other endpoint by changing request path
// before a handler is called. In case of implicit redirect, original path is stored in
// Request.Env.AliasFrom
func (r *Router) Alias(from, to string) *Router {
	return r.Mutator(mutator.Alias(path.Join(r.prefix, from), to))
}
