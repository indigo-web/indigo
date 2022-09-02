package inbuilt

import (
	"github.com/fakefloordiv/indigo/errors"
	"github.com/fakefloordiv/indigo/http/status"
	"github.com/fakefloordiv/indigo/types"
)

/*
This file contains core-callbacks that are called by server, so it's
like a core of the router
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
		return d.renderer.Response(request.Proto, defaultNotFound, respWriter)
	}

	handler, found := urlMethods[request.Method]
	if !found {
		return d.renderer.Response(request.Proto, defaultMethodNotAllowed, respWriter)
	}

	return d.renderer.Response(request.Proto, handler.fun(request), respWriter)
}

// OnError receives error and decides, which error handler is better to use in this case
func (d DefaultRouter) OnError(request *types.Request, respWriter types.ResponseWriter, err error) {
	var code status.Code

	switch err {
	case errors.ErrCloseConnection:
		code = status.ConnectionClose
	case errors.ErrBadRequest:
		code = status.BadRequest
	case errors.ErrTooLarge:
		code = status.RequestEntityTooLarge
	case errors.ErrHeaderFieldsTooLarge:
		code = status.RequestHeaderFieldsTooLarge
	case errors.ErrURITooLong:
		code = status.RequestURITooLong
	case errors.ErrUnsupportedProtocol:
		code = status.HTTPVersionNotSupported
	default:
		// unknown error, but for consistent behaviour we must respond with
		// something. Let it be some neutral error
		code = status.BadRequest
	}

	response := d.errHandlers[code](request)
	_ = d.renderer.Response(request.Proto, response, respWriter)
}
