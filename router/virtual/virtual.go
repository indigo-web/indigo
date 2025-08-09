package virtual

import (
	"github.com/indigo-web/indigo/http"
	"github.com/indigo-web/indigo/http/status"
	"github.com/indigo-web/indigo/internal/strutil"
	"github.com/indigo-web/indigo/router"
	"github.com/indigo-web/indigo/router/virtual/internal/domain"
)

type virtualFabric struct {
	Domain string
	Router router.Builder
}

var _ router.Builder = new(Router)

type Router struct {
	routers       []virtualFabric
	defaultRouter router.Builder
}

// New returns a new instance of the virtual Router
func New() *Router {
	return &Router{}
}

// Host adds a new virtual router. If 0.0.0.0 is passed,
// the router will be set as a default one
func (r *Router) Host(host string, other router.Builder) *Router {
	host = domain.Normalize(host)
	if domain.TrimPort(host) == "0.0.0.0" {
		return r.Default(other)
	}

	r.routers = append(r.routers, virtualFabric{
		Domain: host,
		Router: other,
	})
	return r
}

// Default sets the default router to route requests, Host header value of which aren't
// matched.
// Note: only requests with 0 or 1 Host header values may be passed into the default router.
// If there are more than 1 value, the request will be refused
func (r *Router) Default(def router.Builder) *Router {
	r.defaultRouter = def
	return r
}

func (r *Router) Build() router.Router {
	routers := make([]virtualRouter, len(r.routers))
	for i, fabric := range r.routers {
		routers[i] = virtualRouter{
			Domain: fabric.Domain,
			Router: fabric.Router.Build(),
		}
	}

	var defaultRouter router.Router
	if r.defaultRouter != nil {
		defaultRouter = r.defaultRouter.Build()
	}

	return &runtimeRouter{
		routers:       routers,
		defaultRouter: defaultRouter,
	}
}

type virtualRouter struct {
	Domain string
	Router router.Router
}

var _ router.Router = new(runtimeRouter)

type runtimeRouter struct {
	routers       []virtualRouter
	defaultRouter router.Router
}

func (r *runtimeRouter) OnRequest(request *http.Request) *http.Response {
	virtRouter, resp := r.getRouter(request)
	if virtRouter == nil {
		return resp
	}

	return virtRouter.OnRequest(request)
}

func (r *runtimeRouter) OnError(request *http.Request, err error) *http.Response {
	virtRouter, resp := r.getRouter(request)
	if virtRouter == nil {
		return resp
	}

	return virtRouter.OnError(request, err)
}

// getRouter looks up for the matching router, according to the Host header value.
// If no matching router found, response with already set status code to error is returned.
// If the request is misdirected, default router is returned. Note: it's being returned with
// the error response all together. So be careful to always check the nilness of the returned
// router first
func (r *runtimeRouter) getRouter(request *http.Request) (router.Router, *http.Response) {
	host, found := request.Headers.Lookup("Host")
	if !found {
		return r.defaultRouter, http.Code(request, status.BadRequest)
	}

	for _, virtRouter := range r.routers {
		if strutil.CmpFoldSafe(virtRouter.Domain, host) {
			return virtRouter.Router, nil
		}
	}

	return r.defaultRouter, http.Code(request, status.MisdirectedRequest)
}
