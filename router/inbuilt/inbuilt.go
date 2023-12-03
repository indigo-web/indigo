package inbuilt

import (
	"github.com/indigo-web/indigo/http"
	"github.com/indigo-web/indigo/http/method"
	"github.com/indigo-web/indigo/http/status"
	"github.com/indigo-web/indigo/router"
	"github.com/indigo-web/indigo/router/inbuilt/internal/radix"
	"sort"
	"strings"
)

var _ router.Router = &Router{}

// Router is a built-in implementation of router.Router interface that provides
// some basic router features like middlewares, groups, dynamic routing, error
// handlers, and some implicit things like calling GET-handlers for HEAD-requests,
// or rendering TRACE-responses automatically in case no handler is registered
type Router struct {
	prefix      string
	middlewares []Middleware
	registrar   *registrar
	routesMap   RoutesMap
	tree        radix.Tree
	isStatic    bool
	catchers    []Catcher
	errHandlers errorHandlers

	children  []*Router
	traceBuff []byte
}

// New constructs a new instance of inbuilt router
func New() *Router {
	r := &Router{
		registrar:   newRegistrar(),
		errHandlers: newErrorHandlers(),
	}

	return r
}

// OnStart composes all the registered handlers with middlewares
func (r *Router) OnStart() error {
	r.applyErrorHandlersMiddlewares()

	if err := r.prepare(); err != nil {
		return err
	}

	sort.Slice(r.catchers, func(i, j int) bool {
		return len(r.catchers[i].Prefix) > len(r.catchers[j].Prefix)
	})

	if r.registrar.IsDynamic() {
		r.tree = r.registrar.AsRadixTree()
	} else {
		r.routesMap = r.registrar.AsMap()
		r.isStatic = true
	}

	return nil
}

// OnRequest routes the request
func (r *Router) OnRequest(request *http.Request) *http.Response {
	request.Path = stripTrailingSlash(request.Path)

	if r.isStatic {
		endpoint, found := r.routesMap[request.Path]
		if !found {
			return r.OnError(request, status.ErrNotFound)
		}

		handler := getHandler(request.Method, endpoint.methodsMap)
		if handler == nil {
			request.Env.AllowMethods = endpoint.allow

			return r.OnError(request, status.ErrMethodNotAllowed)
		}

		return handler(request)
	}

	endpoint := r.tree.Match(request.Params, request.Path)
	if endpoint == nil {
		return r.OnError(request, status.ErrNotFound)
	}

	handler := getHandler(request.Method, endpoint.MethodsMap)
	if handler == nil {
		request.Env.AllowMethods = endpoint.Allow

		return r.OnError(request, status.ErrMethodNotAllowed)
	}

	return handler(request)
}

// OnError tries to find a handler for the error, in case it can't - simply
// request.Respond().WithError(...) will be returned
func (r *Router) OnError(request *http.Request, err error) *http.Response {
	if request.Method == method.TRACE && err == status.ErrMethodNotAllowed {
		r.traceBuff = renderHTTPRequest(request, r.traceBuff)

		return traceResponse(request.Respond(), r.traceBuff)
	}

	if err == status.ErrNotFound {
		for _, catcher := range r.catchers {
			if strings.HasPrefix(request.Path, catcher.Prefix) {
				return catcher.Handler(request)
			}
		}
	}

	httpErr, ok := err.(status.HTTPError)
	if !ok {
		return http.Code(request, status.InternalServerError)
	}

	handler := r.retrieveErrorHandler(httpErr.Code)
	if handler == nil {
		// not using http.Error(request, err) for performance purposes, as in this case
		// it would try under the hood to unwrap the error again, however we did this already
		return request.Respond().
			Code(httpErr.Code).
			String(httpErr.Message)
	}

	request.Env.Error = err

	return handler(request)
}

func (r *Router) retrieveErrorHandler(code status.Code) Handler {
	handler, found := r.errHandlers[code]
	if !found {
		return r.errHandlers[AllErrors]
	}

	return handler
}

func (r *Router) applyErrorHandlersMiddlewares() {
	for code, handler := range r.errHandlers {
		r.errHandlers[code] = compose(handler, r.middlewares)
	}
}

// stripTrailingSlash just removes a trailing slash of request path in case it is presented.
// Note: this removes only one trailing slash. In case 2 or more are presented they'll be treated
// as an ordinary part of the path so won't be stripped
func stripTrailingSlash(path string) string {
	if len(path) == 1 {
		return path
	}

	for i := len(path) - 1; i > 1; i-- {
		if path[i] != '/' {
			return path[:i+1]
		}
	}

	return path[0:1]
}

// getHandler looks up for a handler in the methodsMap. In case request method is HEAD, however
// no matching handler is found, a handler for corresponding GET request will be retrieved
func getHandler(reqMethod method.Method, methodsMap MethodsMap) Handler {
	handler := methodsMap[reqMethod]
	if handler == nil && reqMethod == method.HEAD {
		return getHandler(method.GET, methodsMap)
	}

	return handler
}
