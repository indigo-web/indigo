package inbuilt

import (
	"github.com/fakefloordiv/indigo/http"
	"github.com/fakefloordiv/indigo/http/encodings"
	"github.com/fakefloordiv/indigo/types"
)

/*
This file contains core-callbacks that are called by server core.

Methods listed here MUST NOT be called by user ever
*/

// OnStart applies default headers and composes all the registered handlers with middlewares
func (d DefaultRouter) OnStart() {
	d.applyDefaultHeaders()
	d.applyGroups()
	d.applyMiddlewares()
}

// OnRequest routes the request
func (d DefaultRouter) OnRequest(request *types.Request, respWriter types.ResponseWriter) error {
	urlMethods, found := d.routes[request.Path]
	if !found {
		response := d.errHandlers[http.ErrNotFound](request)

		return d.renderer.Response(request, response, respWriter)
	}

	handler, found := urlMethods[request.Method]
	if !found {
		response := d.errHandlers[http.ErrMethodNotAllowed](request)

		return d.renderer.Response(request, response, respWriter)
	}

	return d.renderer.Response(request, handler.fun(request), respWriter)
}

// OnError receives error and decides, which error handler is better to use in this case
func (d DefaultRouter) OnError(request *types.Request, respWriter types.ResponseWriter, err error) {
	response := d.errHandlers[err](request)
	_ = d.renderer.Response(request, response, respWriter)
}

// GetContentEncodings returns encodings.ContentEncodings object as it is
func (d DefaultRouter) GetContentEncodings() encodings.ContentEncodings {
	return d.codings
}
