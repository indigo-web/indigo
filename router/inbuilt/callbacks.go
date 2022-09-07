package inbuilt

import (
	"github.com/fakefloordiv/indigo/http"
	methods "github.com/fakefloordiv/indigo/http/method"
	"github.com/fakefloordiv/indigo/types"
)

/*
This file contains core-callbacks that are called by server core.

Methods listed here MUST NOT be called by user ever
*/

// OnStart composes all the registered handlers with middlewares
func (d DefaultRouter) OnStart() {
	d.applyGroups()
	d.applyMiddlewares()
}

// OnRequest routes the request
func (d DefaultRouter) OnRequest(request *types.Request, render types.Render) error {
	return render(d.processRequest(request))
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
func (d DefaultRouter) OnError(request *types.Request, render types.Render, err error) {
	response := d.errHandlers[err](request)
	_ = render(response)
}
