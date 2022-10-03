package inbuilt

import (
	"context"

	"github.com/fakefloordiv/indigo/http"
	methods "github.com/fakefloordiv/indigo/http/method"
	"github.com/fakefloordiv/indigo/router/inbuilt/obtainer"
	"github.com/fakefloordiv/indigo/types"
	"github.com/fakefloordiv/indigo/valuectx"
)

/*
This file contains core-callbacks that are called by server core.

Methods listed here MUST NOT be called by user ever
*/

// OnStart composes all the registered handlers with middlewares
func (r *Router) OnStart() {
	r.applyGroups()
	r.applyMiddlewares()

	r.obtainer = obtainer.Auto(r.routes)
}

// OnRequest routes the request
func (r *Router) OnRequest(request *types.Request) types.Response {
	return r.processRequest(request)
}

func (r *Router) processRequest(request *types.Request) types.Response {
	ctx, handler, err := r.obtainer(context.Background(), request)
	if err != nil {
		return r.processError(ctx, request, err)
	}

	return handler(ctx, request)
}

// OnError receives an error and calls a corresponding handler. Handler MUST BE
// registered, otherwise panic is raised.
// Luckily (for user), we have all the default handlers registered
func (r *Router) OnError(request *types.Request, err error) types.Response {
	return r.processError(context.Background(), request, err)
}

func (r *Router) processError(ctx context.Context, request *types.Request, err error) types.Response {
	if request.Method == methods.TRACE && err == http.ErrMethodNotAllowed {
		r.traceBuff = renderHTTPRequest(request, r.traceBuff)

		return traceResponse(r.traceBuff)
	}

	handler, found := r.errHandlers[err]
	if !found {
		return types.WithError(err)
	}

	ctx = valuectx.WithValue(ctx, "error", err)

	return handler(ctx, request)
}
