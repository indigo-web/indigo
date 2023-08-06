package inbuilt

import (
	"github.com/indigo-web/indigo/ctx"
	"github.com/indigo-web/indigo/http"
	"github.com/indigo-web/indigo/http/method"
	"github.com/indigo-web/indigo/http/status"
	"github.com/indigo-web/indigo/router"
	"github.com/indigo-web/indigo/router/inbuilt/radix"
	"github.com/indigo-web/indigo/router/inbuilt/rmap"
	"github.com/indigo-web/indigo/router/inbuilt/types"
)

var _ router.Router = &Router{}

// Router is a built-in implementation of router.Router interface that provides
// some basic router features like middlewares, groups, dynamic routing, error
// handlers, and some implicit things like calling GET-handlers for HEAD-requests,
// or rendering TRACE-responses automatically in case no handler is registered
type Router struct {
	prefix           string
	middlewares      []types.Middleware
	registrar        *registrar
	routesMap        *rmap.Map
	tree             radix.Tree
	isStatic         bool
	errHandlers      types.ErrHandlers
	reusableErrCtx   ctx.ReusableContext[string, error]
	reusableAllowCtx ctx.ReusableContext[string, string]

	children  []*Router
	traceBuff []byte
}

// New constructs a new instance of inbuilt router
func New() *Router {
	r := &Router{
		registrar:        newRegistrar(),
		errHandlers:      newErrorHandlers(),
		reusableErrCtx:   ctx.NewReusable[string, error](),
		reusableAllowCtx: ctx.NewReusable[string, string](),
	}

	return r
}

// OnStart composes all the registered handlers with middlewares
func (r *Router) OnStart() error {
	if err := r.prepare(); err != nil {
		return err
	}

	if r.registrar.IsDynamic() {
		r.tree = r.registrar.AsRadixTree()
	} else {
		r.routesMap = r.registrar.AsRMap()
		r.isStatic = true
	}

	return nil
}

// OnRequest routes the request
func (r *Router) OnRequest(request *http.Request) http.Response {
	request.Path.String = stripTrailingSlash(request.Path.String)

	if r.isStatic {
		methodsMap, allow, ok := r.routesMap.Get(request.Path.String)
		if !ok {
			return r.OnError(request, status.ErrNotFound)
		}

		handler := getHandler(request.Method, methodsMap)
		if handler == nil {
			r.reusableAllowCtx.Set(request.Ctx, "allow", allow)
			request.Ctx = r.reusableAllowCtx

			return r.OnError(request, status.ErrMethodNotAllowed)
		}

		return handler(request)
	}

	payload := r.tree.Match(request.Path.Params, request.Path.String)
	if payload == nil {
		return r.OnError(request, status.ErrNotFound)
	}

	handler := getHandler(request.Method, payload.MethodsMap)
	if handler == nil {
		r.reusableAllowCtx.Set(request.Ctx, "allow", payload.Allow)
		request.Ctx = r.reusableAllowCtx

		return r.OnError(request, status.ErrMethodNotAllowed)
	}

	return handler(request)
}

// OnError tries to find a handler for the error, in case it can't - simply
// request.Respond().WithError(...) will be returned
func (r *Router) OnError(request *http.Request, err error) http.Response {
	if request.Method == method.TRACE && err == status.ErrMethodNotAllowed {
		r.traceBuff = renderHTTPRequest(request, r.traceBuff)

		return traceResponse(request.Respond(), r.traceBuff)
	}

	handler, found := r.errHandlers[err]
	if !found {
		return request.Respond().WithError(err)
	}

	r.reusableErrCtx.Set(request.Ctx, "error", err)
	request.Ctx = r.reusableErrCtx

	return handler(request)
}

// stripTrailingSlash just removes a trailing slash of request path in case it is presented.
// Note: this removes only one trailing slash. In case 2 or more are presented they'll be treated
// as an ordinary part of the path so won't be stripped
func stripTrailingSlash(path string) string {
	if path[len(path)-1] == '/' && len(path) > 1 {
		return path[:len(path)-1]
	}

	return path
}

// getHandler looks for a handler in the methodsMap. In case not found, it checks whether
// the method is HEAD. In this case, we're looking for a GET method handler, as semantically
// both methods are same, except response body (response to a HEAD request MUST NOT contain
// a body)
func getHandler(reqMethod method.Method, methodsMap types.MethodsMap) types.Handler {
	handler := methodsMap[reqMethod]
	if handler == nil {
		if reqMethod == method.HEAD {
			return getHandler(method.GET, methodsMap)
		}

		return nil
	}

	return handler
}
