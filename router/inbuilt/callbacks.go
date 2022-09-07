package inbuilt

import (
	"github.com/fakefloordiv/indigo/http"
	"github.com/fakefloordiv/indigo/http/headers"
	methods "github.com/fakefloordiv/indigo/http/method"
	"github.com/fakefloordiv/indigo/http/proto"
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
		if request.Method == methods.TRACE {
			return renderRequest(request)
		}

		return d.errHandlers[http.ErrNotFound](request)
	}

	handler, found := urlMethods[request.Method]
	switch found {
	case true:
		return handler.fun(request)
	default:
		switch request.Method {
		case methods.HEAD:
			// by default, if no handler for HEAD method is registered, automatically
			// call a corresponding GET method - renderer anyway will discard request
			// body and leave only response line with headers, just like rfc2068, 9.4
			// wants
			handler, found = urlMethods[methods.GET]
			if found {
				return handler.fun(request)
			}
		case methods.TRACE:
			return renderRequest(request)
		}

		return d.errHandlers[http.ErrMethodNotAllowed](request)
	}
}

// OnError receives an error and calls a corresponding handler. Handler MUST BE
// registered, otherwise panic is raised.
// Luckily (for user), we have all the default handlers registered
func (d DefaultRouter) OnError(request *types.Request, render types.Render, err error) {
	response := d.errHandlers[err](request)
	_ = render(response)
}

func renderRequest(request *types.Request) types.Response {
	requestLine := types.WithResponse.
		WithHeader("Content-Type", "message/http").
		WithBody(methods.ToString(request.Method) + http.SP).
		WithBodyAppend(request.Path + http.SP).
		WithBodyByteAppend(proto.ToBytes(request.Proto)).
		WithBodyByteAppend(http.CRLF)

	return renderHeadersInto(request.Headers, requestLine).WithBodyByteAppend(http.CRLF)
}

func renderHeadersInto(headers headers.Headers, response types.Response) types.Response {
	for k, v := range headers {
		response = response.WithBodyAppend(k).WithBodyAppend(http.COLON).WithBodyAppend(http.SP)

		for i := 0; i < len(v)-1; i++ {
			response = response.WithBodyAppend(v[i] + http.COMMA)
		}

		response = response.WithBodyAppend(v[len(v)-1]).WithBodyByteAppend(http.CRLF)
	}

	return response
}
