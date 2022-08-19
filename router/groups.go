package router

/*
This file is responsible for endpoint groups
*/

// Group creates a new instance of DefaultRouter, but inherited from current one
// Middlewares has to be inherited from a parent, but adding new middlewares
// in a child group MUST NOT affect parent ones, so parent middlewares
// are copied into child ones. Everything else is inherited from parent as it is
func (d DefaultRouter) Group(prefix string) DefaultRouter {
	var newMiddlewares []Middleware

	return DefaultRouter{
		prefix:      d.prefix + prefix,
		middlewares: append(newMiddlewares, d.middlewares...),
		routes:      d.routes,
		errHandlers: d.errHandlers,
		renderer:    d.renderer,
	}
}
