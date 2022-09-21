package inbuilt

import (
	"context"
	context2 "github.com/fakefloordiv/indigo/valuectx"
	"strings"

	methods "github.com/fakefloordiv/indigo/http/method"
	"github.com/fakefloordiv/indigo/internal/mapconv"
	"github.com/fakefloordiv/indigo/types"
)

/*
This file contains core-callbacks that are called by server core.

Methods listed here MUST NOT be called by user ever
*/

// OnStart composes all the registered handlers with middlewares
func (r *Router) OnStart() {
	r.requestProcessor = r.staticProcessor
	r.applyGroups()
	r.applyMiddlewares()
	r.loadAllowedMethods()
}

// OnRequest routes the request
func (r *Router) OnRequest(request *types.Request, render types.Render) error {
	return render(r.requestProcessor(request))
}

// OnError receives an error and calls a corresponding handler. Handler MUST BE
// registered, otherwise panic is raised.
// Luckily (for user), we have all the default handlers registered
func (r Router) OnError(request *types.Request, render types.Render, err error) {
	_ = render(r.processError(context.Background(), request, err))
}

func (r Router) processError(ctx context.Context, request *types.Request, err error) types.Response {
	handler, found := r.errHandlers[err]
	if !found {
		return types.WithResponse.WithError(err)
	}

	ctx = context2.WithValue(ctx, "error", err)

	return handler(ctx, request)
}

func (r Router) loadAllowedMethods() {
	for k, v := range r.routes {
		allowedMethods := mapconv.Keys[methods.Method, *handlerObject](v)
		r.allowedMethods[k] = strings.Join(methods2string(allowedMethods...), ",")
	}
}

func methods2string(ms ...methods.Method) []string {
	out := make([]string, 0, len(ms))

	for _, method := range ms {
		out = append(out, methods.ToString(method))
	}

	return out
}
