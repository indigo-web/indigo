package router

import (
	"indigo/errors"
	"indigo/http/status"
	"indigo/types"
)

func (d DefaultRouter) OnRequest(request *types.Request, respWriter types.ResponseWriter) error {
	urlMethods, found := d.routes[request.Path]
	if !found {
		return respWriter(d.renderer.Response(request.Proto, defaultNotFound))
	}

	handler, found := urlMethods[request.Method]
	if !found {
		return respWriter(d.renderer.Response(request.Proto, defaultMethodNotAllowed))
	}

	return respWriter(d.renderer.Response(request.Proto, handler(request)))
}

func (d DefaultRouter) OnError(request *types.Request, respWriter types.ResponseWriter, err error) {
	var code status.Code

	switch err {
	case errors.ErrCloseConnection:
		code = status.ConnectionClose
	case errors.ErrBadRequest:
		code = status.BadRequest
	case errors.ErrTooLarge:
		code = status.RequestEntityTooLarge
	default:
		// unknown error, but for consistent behaviour we must respond with
		// something. Let it be some neutral error
		code = status.BadRequest
	}

	response := d.errHandlers[code](request)
	_ = respWriter(d.renderer.Response(request.Proto, response))
}
