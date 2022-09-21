package inbuilt

import (
	"context"
	"strings"

	"github.com/fakefloordiv/indigo/http"
	methods "github.com/fakefloordiv/indigo/http/method"
	context2 "github.com/fakefloordiv/indigo/internal/context"
	"github.com/fakefloordiv/indigo/internal/mapconv"
	"github.com/fakefloordiv/indigo/types"
)

/*
This file contains core-callbacks that are called by server core.

Methods listed here MUST NOT be called by user ever
*/

// OnStart composes all the registered handlers with middlewares
func (d Router) OnStart() {
	d.applyGroups()
	d.applyMiddlewares()
	d.loadAllowedMethods()
}

// OnRequest routes the request
func (d *Router) OnRequest(request *types.Request, render types.Render) error {
	return render(d.processRequest(request))
}

func (d *Router) processRequest(request *types.Request) types.Response {
	ctx := context.Background()

	urlMethods, found := d.routes[request.Path]
	if !found {
		if request.Method == methods.TRACE {
			d.traceBuff = renderHTTPRequest(request, d.traceBuff)

			return traceResponse(d.traceBuff)
		}

		return d.processError(ctx, request, http.ErrNotFound)
	}

	handler, found := urlMethods[request.Method]
	switch found {
	case true:
		return handler.fun(context.Background(), request)
	default:
		switch request.Method {
		case methods.HEAD:
			// by default, if no handler for HEAD method is registered, automatically
			// call a corresponding GET method - renderer anyway will discard request
			// body and leave only response line with headers, just like rfc2068, 9.4
			// wants
			handler, found = urlMethods[methods.GET]
			if found {
				return handler.fun(ctx, request)
			}
		case methods.TRACE:
			d.traceBuff = renderHTTPRequest(request, d.traceBuff)

			return traceResponse(d.traceBuff)
		}

		ctx = context2.WithValue(ctx, "allow", d.allowedMethods[request.Path])

		return d.processError(ctx, request, http.ErrMethodNotAllowed)
	}
}

// OnError receives an error and calls a corresponding handler. Handler MUST BE
// registered, otherwise panic is raised.
// Luckily (for user), we have all the default handlers registered
func (d Router) OnError(request *types.Request, render types.Render, err error) {
	_ = render(d.processError(context.Background(), request, err))
}

func (d Router) processError(ctx context.Context, request *types.Request, err error) types.Response {
	handler, found := d.errHandlers[err]
	if !found {
		return types.WithResponse.WithError(err)
	}

	ctx = context2.WithValue(ctx, "error", err)

	return handler(ctx, request)
}

func (d Router) loadAllowedMethods() {
	for k, v := range d.routes {
		allowedMethods := mapconv.Keys[methods.Method, *handlerObject](v)
		d.allowedMethods[k] = strings.Join(methods2string(allowedMethods...), ",")
	}
}

func methods2string(ms ...methods.Method) []string {
	out := make([]string, 0, len(ms))

	for _, method := range ms {
		out = append(out, methods.ToString(method))
	}

	return out
}
