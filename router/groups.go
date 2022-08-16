package router

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
