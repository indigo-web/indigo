package inbuilt

import (
	"github.com/fakefloordiv/indigo/http"
	"github.com/fakefloordiv/indigo/http/headers"
	methods "github.com/fakefloordiv/indigo/http/method"
	"github.com/fakefloordiv/indigo/types"
)

/*
This file contains core-callbacks that are called by server core.

Methods listed here MUST NOT be called by user ever
*/

// OnStart applies default headers and composes all the registered handlers with middlewares
// and sets passed default headers from application to renderer
func (d DefaultRouter) OnStart(defaultHeaders headers.Headers) {
	d.renderer.SetDefaultHeaders(defaultHeaders)
	d.applyGroups()
	d.applyMiddlewares()
}

// OnRequest routes the request
func (d DefaultRouter) OnRequest(request *types.Request, respWriter types.ResponseWriter) error {
	return d.renderer.Response(request, d.processRequest(request), respWriter)
}

func (d DefaultRouter) processRequest(request *types.Request) types.Response {
	urlMethods, found := d.routes[request.Path]
	if !found {
		return d.errHandlers[http.ErrNotFound](request)
	}

	handler, found := urlMethods[request.Method]
	switch found {
	case true:
		return handler.fun(request)
	default:
		// by default, if no handler for HEAD method is registered, automatically
		// call a corresponding GET method - renderer anyway will discard request
		// body and leave only response line with headers, just like rfc2068, 9.4
		// wants
		if request.Method == methods.HEAD {
			handler, found = urlMethods[methods.GET]
			if found {
				return handler.fun(request)
			}
		}

		return d.errHandlers[http.ErrMethodNotAllowed](request)
	}
}

// OnError receives error and decides, which error handler is better to use in this case
func (d DefaultRouter) OnError(request *types.Request, respWriter types.ResponseWriter, err error) {
	response := d.errHandlers[err](request)
	_ = d.renderer.Response(request, response, respWriter)
}
