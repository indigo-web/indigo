package inbuilt

import (
	"github.com/indigo-web/indigo/router/inbuilt/internal/types"
)

// Mutator is kind of pre-middleware. It's being called at the moment, when a request arrives
// to the router, but before the routing will be done. So by that, the request may be mutated.
// For example, mutator may normalize requests' paths, log them, make invisible redirects, etc.
type Mutator = types.Mutator

// Mutator adds a new mutator.
//
// NOTE: registering them on groups will affect only the order of execution
func (r *Router) Mutator(mutator Mutator) *Router {
	r.mutators = append(r.mutators, mutator)
	return r
}
