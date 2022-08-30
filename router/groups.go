package router

/*
This file is responsible for endpoint groups
*/

// Group creates a new instance of DefaultRouter, but inherited from current one
// Middlewares has to be inherited from a parent, but adding new middlewares
// in a child group MUST NOT affect parent ones, so parent middlewares
// are copied into child ones. Everything else is inherited from parent as it is
func (d DefaultRouter) Group(prefix string) *DefaultRouter {
	var newMiddlewares []Middleware

	r := &DefaultRouter{
		root:        d.root,
		prefix:      d.prefix + prefix,
		middlewares: append(newMiddlewares, d.middlewares...),
		routes:      make(routesMap),
		errHandlers: d.errHandlers,
		renderer:    d.renderer,
		codings:     d.codings,
	}

	d.root.groups = append(d.root.groups, *r)

	return r
}

func (d DefaultRouter) applyGroups() {
	for _, group := range d.groups {
		mergeRoutes(d.routes, group.routes)
	}
}

func mergeRoutes(into, values routesMap) {
	for key, value := range values {
		into[key] = value
	}
}
