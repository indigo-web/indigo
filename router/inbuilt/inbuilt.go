package inbuilt

import (
	"path"

	"github.com/indigo-web/indigo/http"
	"github.com/indigo-web/indigo/http/method"
	"github.com/indigo-web/indigo/http/status"
	"github.com/indigo-web/indigo/router"
	"github.com/indigo-web/indigo/router/inbuilt/internal"
	"github.com/indigo-web/indigo/router/inbuilt/mutator"
	"github.com/indigo-web/indigo/router/inbuilt/uri"
)

// Middleware works like a chain of nested calls, next may be even directly
// handler. But if we are not a closing middleware, we will call next
// middleware that is simply a partial middleware with already provided next
type Middleware func(next Handler, request *http.Request) *http.Response

var _ router.Builder = new(Router)

// Router is a recommended router for indigo. It features groups, middlewares, pre-middlewares,
// resources, automatic OPTIONS and TRACE response capabilities and dynamic routing (enabled
// automatically if any of routes is dynamic, otherwise more efficient map-based static routing
// is used.)
type Router struct {
	enableTRACE  bool
	prefix       string
	mutators     []Mutator
	middlewares  []Middleware
	registrar    *registrar
	children     []*Router
	traceHandler Handler
	errHandlers  errorHandlers
}

// New constructs a new instance of inbuilt router
func New() *Router {
	return &Router{
		registrar:   newRegistrar(),
		errHandlers: newErrorHandlers(),
	}
}

// AllErrors tells the Router.RouteError to use the passed error handler as a generic
// handler. A generic error handler is usually called only if no other was matched.
const AllErrors = status.Code(0)

// Route registers a new endpoint.
func (r *Router) Route(method method.Method, path string, handler Handler, middlewares ...Middleware) *Router {
	err := r.registrar.Add(r.prefix+path, method, compose(handler, middlewares))
	if err != nil {
		panic(err)
	}

	return r
}

// TODO: update the error handling mechanism. It should be more modifications-prone

// RouteError adds an error handler for a corresponding HTTP error code.
//
// The following error codes may be registered:
//   - AllErrors (called only if no other error handlers found)
//   - status.BadRequest
//   - status.NotFound
//   - status.MethodNotAllowed
//   - status.RequestEntityTooLarge
//   - status.CloseConnection
//   - status.RequestURITooLong
//   - status.HeaderFieldsTooLarge
//   - status.HTTPVersionNotSupported
//   - status.UnsupportedMediaType
//   - status.NotImplemented
//   - status.RequestTimeout
//
// Note: if handler returned one of error codes above, error handler WON'T be called.
// Also, global middlewares, applied to the root router, will also be used for error handlers.
// However, global middlewares defined on groups won't be used.
//
// WARNING: calling this method from groups will affect ALL routers, including root
func (r *Router) RouteError(handler Handler, codes ...status.Code) *Router {
	if len(codes) == 0 {
		codes = append(codes, AllErrors)
	}

	for _, code := range codes {
		r.errHandlers[code] = handler
	}

	return r
}

// Use registers a new middleware in the group.
func (r *Router) Use(middlewares ...Middleware) *Router {
	r.middlewares = append(r.middlewares, middlewares...)
	return r
}

func (r *Router) applyMiddlewares() {
	r.registrar.Apply(func(handler Handler) Handler {
		return compose(handler, r.middlewares)
	})
}

// Group creates a subrouter with its own scoping and path prefix. The scoping affects mainly
// middleware application rules: a new group inherits its parental middlewares, but middlewares,
// registered on the group, don't affect its parents ones. Parent middlewares are chained first,
// therefore will also be called earlier than middlewares registered directly on the group.
func (r *Router) Group(prefix string) *Router {
	subrouter := &Router{
		prefix:      r.prefix + prefix,
		registrar:   newRegistrar(),
		errHandlers: r.errHandlers,
	}

	r.children = append(r.children, subrouter)

	return subrouter
}

func (r *Router) prepare() error {
	for _, child := range r.children {
		if err := child.prepare(); err != nil {
			return err
		}

		if err := r.registrar.Merge(child.registrar); err != nil {
			return err
		}

		r.mutators = append(r.mutators, child.mutators...)
	}

	r.applyMiddlewares()

	return nil
}

// Resource returns a new Resource object for a provided resource path.
func (r *Router) Resource(path string) Resource {
	return Resource{
		group: r.Group(path),
	}
}

// Alias is an implicit redirect, made absolutely transparently before a specific handler is chosen.
// The original path is stored in Request.Env.AliasFrom. Optionally only specific methods can be set
// to be aliased. Otherwise, ANY requests matching alias will be aliased, which might not always be
// the desired behavior.
func (r *Router) Alias(from, to string, forMethods ...method.Method) *Router {
	return r.Mutator(mutator.Alias(path.Join(r.prefix, from), to, forMethods...))
}

type Mutator = internal.Mutator

// Mutator adds a new Mutator. Please note that groups scoping rules don't apply on them, only the
// execution order is affected.
func (r *Router) Mutator(mutator Mutator) *Router {
	r.mutators = append(r.mutators, mutator)
	return r
}

// EnableTRACE allows the router to automatically respond to TRACE requests if there is no
// matching handler registered. To explore why it's better to keep the option disabled, see
// https://owasp.org/www-community/attacks/Cross_Site_Tracing
func (r *Router) EnableTRACE(flag bool) *Router {
	r.enableTRACE = flag
	return r
}

// runtimeRouter is a compiled router. Router represents a "dummy" builder, while the actual
// action happens here.
type runtimeRouter struct {
	enableTRACE   bool
	isStatic      bool
	tree          radixTree
	routesMap     routesMap
	errHandlers   errorHandlers
	serverOptions string
	mutators      []Mutator
}

func (r *Router) Build() router.Router {
	r.applyErrorHandlersMiddlewares()

	if err := r.prepare(); err != nil {
		panic(err)
	}

	isDynamic := r.registrar.IsDynamic()
	var (
		rmap routesMap
		tree radixTree
	)
	if isDynamic {
		tree = r.registrar.AsRadixTree()
	} else {
		rmap = r.registrar.AsMap()
	}

	return &runtimeRouter{
		enableTRACE:   r.enableTRACE,
		isStatic:      !isDynamic,
		tree:          tree,
		routesMap:     rmap,
		errHandlers:   r.errHandlers,
		serverOptions: r.registrar.Options(r.enableTRACE),
		mutators:      r.mutators,
	}
}

// OnRequest processes the request
func (r *runtimeRouter) OnRequest(request *http.Request) *http.Response {
	request.Path = uri.Normalize(request.Path)
	r.runMutators(request)

	return r.onRequest(request)
}

func (r *runtimeRouter) onRequest(request *http.Request) *http.Response {
	var (
		e     endpoint
		found bool
	)

	if r.isStatic {
		e, found = r.routesMap[request.Path]
	} else {
		e, found = r.tree.Lookup(request.Path, request.Vars)
	}

	if !found {
		return r.onError(request, status.ErrNotFound)
	}

	handler := getHandler(request.Method, e.methods)
	if handler == nil {
		request.Env.AllowedMethods = e.allow

		return r.onError(request, status.ErrMethodNotAllowed)
	}

	return handler(request)
}

// OnError uses a user-defined error handler, otherwise default http.Error
func (r *runtimeRouter) OnError(request *http.Request, err error) *http.Response {
	r.runMutators(request)

	return r.onError(request, err)
}

func (r *runtimeRouter) onError(request *http.Request, err error) *http.Response {
	if request.Method == method.TRACE && err == status.ErrMethodNotAllowed {
		r.traceBuff = renderHTTPRequest(request, r.traceBuff)

		return traceResponse(request.Respond(), r.traceBuff)
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

func (r *runtimeRouter) runMutators(request *http.Request) {
	for _, mutator := range r.mutators {
		mutator(request)
	}
}

func (r *runtimeRouter) retrieveErrorHandler(code status.Code) Handler {
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

// compose produces an array of middlewares into the chain, represented by types.Handler
func compose(handler Handler, middlewares []Middleware) Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		handler = func(handler Handler, middleware Middleware) Handler {
			return func(request *http.Request) *http.Response {
				return middleware(handler, request)
			}
		}(handler, middlewares[i])
	}

	return handler
}

// getHandler looks up for a handler in the methodsMap. In case request method is HEAD, however
// no matching handler is found, a handler for corresponding GET request will be retrieved
func getHandler(reqMethod method.Method, mlut methodLUT) Handler {
	handler := mlut[reqMethod]
	if handler == nil && reqMethod == method.HEAD {
		return getHandler(method.GET, mlut)
	}

	return handler
}

func isServerWideOptions(req *http.Request) bool {
	return req.Method == method.OPTIONS && req.Path == "*"
}
