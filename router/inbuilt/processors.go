package inbuilt

import (
	"context"
	"github.com/fakefloordiv/indigo/http"
	methods "github.com/fakefloordiv/indigo/http/method"
	context2 "github.com/fakefloordiv/indigo/internal/context"
	"github.com/fakefloordiv/indigo/types"
)

/*
This file is responsible for a different implementations of routing. Each
is optimized in its case, and chosen for use on start up
*/

func (r *Router) staticProcessor(request *types.Request) types.Response {
	ctx := context.Background()

	urlMethods, found := r.routes[request.Path]
	if !found {
		if request.Method == methods.TRACE {
			r.traceBuff = renderHTTPRequest(request, r.traceBuff)

			return traceResponse(r.traceBuff)
		}

		return r.processError(ctx, request, http.ErrNotFound)
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
			r.traceBuff = renderHTTPRequest(request, r.traceBuff)

			return traceResponse(r.traceBuff)
		}

		ctx = context2.WithValue(ctx, "allow", r.allowedMethods[request.Path])

		return r.processError(ctx, request, http.ErrMethodNotAllowed)
	}
}
