package inbuilt

import (
	"github.com/indigo-web/indigo/router/inbuilt/internal"
)

type Mutator = internal.Mutator

// Mutator adds a new mutator.
//
// NOTE: registering them on groups will affect only the order of execution
func (r *Router) Mutator(mutator Mutator) *Router {
	r.mutators = append(r.mutators, mutator)
	return r
}
