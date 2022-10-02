package inbuilt

import (
	routertypes "github.com/fakefloordiv/indigo/router/inbuilt/types"
)

/*
This file is responsible for endpoint groups
*/

// Group creates a new instance of InbuiltRouter, but inherited from current one
// Middlewares has to be inherited from a parent, but adding new middlewares
// in a child group MUST NOT affect parent ones, so parent middlewares
// are copied into child ones. Everything else is inherited from parent as it is
func (r Router) Group(prefix string) *Router {
	var newMiddlewares []routertypes.Middleware

	router := &Router{
		root:        r.root,
		prefix:      r.prefix + prefix,
		middlewares: append(newMiddlewares, r.middlewares...),
		routes:      make(routertypes.RoutesMap),
		errHandlers: r.errHandlers,
	}

	r.root.groups = append(r.root.groups, *router)

	return router
}

func (r Router) applyGroups() {
	for _, group := range r.groups {
		mergeRoutes(r.routes, group.routes)
	}
}

func mergeRoutes(into, values routertypes.RoutesMap) {
	for key, value := range values {
		into[key] = value
	}
}
