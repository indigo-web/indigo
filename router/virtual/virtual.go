package virtual

import (
	"github.com/indigo-web/indigo/http"
	"github.com/indigo-web/indigo/http/status"
	"github.com/indigo-web/indigo/router"
	"github.com/indigo-web/indigo/router/virtual/internal/domain"
	"github.com/indigo-web/utils/strcomp"
)

type virtualRouter struct {
	Domain string
	Router router.Router
}

var _ router.Router = new(Router)

type Router struct {
	routers       []virtualRouter
	defaultRouter router.Router
}

// New returns a new instance of the virtual Router
func New() *Router {
	return &Router{}
}

// Host adds a new virtual router. If 0.0.0.0 is passed,
// the router will be set as a default one
func (r *Router) Host(host string, other router.Router) *Router {
	host = domain.Normalize(host)
	if domain.TrimPort(host) == "0.0.0.0" {
		return r.Default(other)
	}

	r.routers = append(r.routers, virtualRouter{
		Domain: host,
		Router: other,
	})
	return r
}

// Default sets the default router to route requests, Host header value of which aren't
// matched.
// Note: only requests with 0 or 1 Host header values may be passed into the default router.
// If there are more than 1 value, the request will be refused
func (r *Router) Default(def router.Router) *Router {
	r.defaultRouter = def
	return r
}

func (r *Router) OnStart() error {
	for _, host := range r.routers {
		if err := host.Router.OnStart(); err != nil {
			return err
		}
	}

	if r.defaultRouter != nil {
		return r.defaultRouter.OnStart()
	}

	return nil
}

func (r *Router) OnRequest(request *http.Request) *http.Response {
	virtRouter, resp := r.getRouter(request)
	if virtRouter == nil {
		return resp
	}

	return virtRouter.OnRequest(request)
}

func (r *Router) OnError(request *http.Request, err error) *http.Response {
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
func (r *Router) getRouter(request *http.Request) (router.Router, *http.Response) {
	hosts := request.Headers.Values("host")
	switch len(hosts) {
	case 0:
		return r.defaultRouter, http.Code(request, status.BadRequest)
	case 1:
	default:
		return nil, http.Code(request, status.BadRequest)
	}

	host := hosts[0]

	for _, virtRouter := range r.routers {
		if strcomp.EqualFold(virtRouter.Domain, host) {
			return virtRouter.Router, nil
		}
	}

	return r.defaultRouter, http.Code(request, status.MisdirectedRequest)
}
