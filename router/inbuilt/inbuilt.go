package inbuilt

import (
	"github.com/indigo-web/indigo/http"
	"github.com/indigo-web/indigo/http/method"
	"github.com/indigo-web/indigo/http/status"
	"github.com/indigo-web/indigo/router"
	"github.com/indigo-web/indigo/router/inbuilt/internal/radix"
	"github.com/indigo-web/indigo/router/inbuilt/internal/types"
	"github.com/indigo-web/indigo/router/inbuilt/uri"
	"sort"
	"strings"
)

var _ router.Router = &Router{}

// Router is a built-in implementation of router.Router interface that provides
// some basic router features like middlewares, groups, dynamic routing, error
// handlers, and some implicit things like calling GET-handlers for HEAD-requests,
// or rendering TRACE-responses automatically in case no handler is registered
type Router struct {
	isRoot      bool
	prefix      string
	mutators    []Mutator
	middlewares []Middleware
	catchers    []Catcher
	registrar   *registrar
	routesMap   routesMap
	tree        radix.Tree
	isStatic    bool
	errHandlers errorHandlers

	children  []*Router
	traceBuff []byte
}

// New constructs a new instance of inbuilt router
func New() *Router {
	r := &Router{
		isRoot:      true,
		registrar:   newRegistrar(),
		errHandlers: newErrorHandlers(),
	}

	return r
}

// OnStart initializes the router. It merges all the groups and prepares
func (r *Router) OnStart() error {
	r.applyErrorHandlersMiddlewares()

	if err := r.prepare(); err != nil {
		return err
	}

	sort.Slice(r.catchers, func(i, j int) bool {
		return len(r.catchers[i].Prefix) > len(r.catchers[j].Prefix)
	})
	r.applyCatchersMiddlewares()

	r.isStatic = !r.registrar.IsDynamic()
	if r.isStatic {
		r.routesMap = r.registrar.AsMap()
	} else {
		r.tree = r.registrar.AsRadixTree()
	}

	return nil
}

// OnRequest processes the request
func (r *Router) OnRequest(request *http.Request) *http.Response {
	r.runMutators(request)

	// TODO: should path normalization be implemented as a mutator?
	request.Path = uri.Normalize(request.Path)

	return r.onRequest(request)
}

func (r *Router) onRequest(request *http.Request) *http.Response {
	var methodsMap types.MethodsMap

	if r.isStatic {
		endpoint, found := r.routesMap[request.Path]
		if !found {
			return r.onError(request, status.ErrNotFound)
		}

		methodsMap = endpoint.methodsMap
		request.Env.AllowMethods = endpoint.allow
	} else {
		endpoint := r.tree.Match(request.Path, request.Params)
		if endpoint == nil {
			return r.onError(request, status.ErrNotFound)
		}

		methodsMap = endpoint.MethodsMap
		request.Env.AllowMethods = endpoint.Allow
	}

	handler := getHandler(request.Method, methodsMap)
	if handler == nil {
		return r.onError(request, status.ErrMethodNotAllowed)
	}

	return handler(request)
}

// OnError tries to find a handler for the error, in case it can't - simply
// request.Respond().WithError(...) will be returned
func (r *Router) OnError(request *http.Request, err error) *http.Response {
	r.runMutators(request)

	return r.onError(request, err)
}

func (r *Router) onError(request *http.Request, err error) *http.Response {
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
		// not using http.Error(request, err) in performance purposes, as in this case
		// it would try under the hood to unwrap the error again, however we did this already
		return request.Respond().
			Code(httpErr.Code).
			String(httpErr.Message)
	}

	request.Env.Error = err

	return handler(request)
}

func (r *Router) runMutators(request *http.Request) {
	for _, mutator := range r.mutators {
		mutator(request)
	}
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

// getHandler looks up for a handler in the methodsMap. In case request method is HEAD, however
// no matching handler is found, a handler for corresponding GET request will be retrieved
func getHandler(reqMethod method.Method, methodsMap types.MethodsMap) Handler {
	handler := methodsMap[reqMethod]
	if handler == nil && reqMethod == method.HEAD {
		return getHandler(method.GET, methodsMap)
	}

	return handler
}
