package inbuilt

import "github.com/indigo-web/indigo/http"

// Alias makes an implicitly redirects to other endpoint by changing request path
// before a handler is called. In case of implicit redirect, original path is stored in
// Request.Env.AliasFrom
func (r *Router) Alias(from, to string) *Router {
	if r.aliases == nil {
		r.aliases = make(map[string]string, 1)
	}

	r.aliases[from] = to
	return r
}

func (r *Router) retrieveAlias(request *http.Request) {
	if to, found := r.aliases[request.Path]; found {
		request.Env.AliasFrom, request.Path = request.Path, to
	}
}
