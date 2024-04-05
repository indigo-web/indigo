package inbuilt

/*
This file is responsible for endpoint groups
*/

// Group creates a new router with pre-defined prefix for all paths. It'll automatically be
// merged into the head router on server start. Middlewares, applied on this router, will not
// affect the head router, but initially head router's middlewares will be inherited and will
// be called in the first order. Registering new error handlers will result in affecting error
// handlers among ALL the existing groups, including head router
func (r *Router) Group(prefix string) *Router {
	subrouter := &Router{
		prefix:      r.prefix + prefix,
		registrar:   newRegistrar(),
		errHandlers: r.errHandlers,
	}

	r.children = append(r.children, subrouter)

	return subrouter
}

func (r *Router) prepare() error {
	for _, child := range r.children {
		if err := child.prepare(); err != nil {
			return err
		}

		if err := r.registrar.Merge(child.registrar); err != nil {
			return err
		}

		r.mutators = append(r.mutators, child.mutators...)
		r.catchers = append(r.catchers, child.catchers...)
	}

	r.applyMiddlewares()
	r.applyCatchersMiddlewares()

	return nil
}
