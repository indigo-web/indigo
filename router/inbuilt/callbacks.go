package inbuilt

import (
	"github.com/indigo-web/indigo/http/status"

	"github.com/indigo-web/indigo/http"
	methods "github.com/indigo-web/indigo/http/method"
	"github.com/indigo-web/indigo/router/inbuilt/obtainer"
	"github.com/indigo-web/indigo/valuectx"
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
func (r *Router) OnRequest(request *http.Request) http.Response {
	return r.processRequest(request)
}

func (r *Router) processRequest(request *http.Request) http.Response {
	handler, err := r.obtainer(request)
	if err != nil {
		return r.processError(request, err)
	}

	return handler(request)
}

// OnError receives an error and calls a corresponding handler. Handler MUST BE
// registered, otherwise panic is raised.
// Luckily (for user), we have all the default handlers registered
func (r *Router) OnError(request *http.Request, err error) http.Response {
	return r.processError(request, err)
}

func (r *Router) processError(request *http.Request, err error) http.Response {
	if request.Method == methods.TRACE && err == status.ErrMethodNotAllowed {
		r.traceBuff = renderHTTPRequest(request, r.traceBuff)

		return traceResponse(http.RespondTo(request), r.traceBuff)
	}

	handler, found := r.errHandlers[err]
	if !found {
		return http.RespondTo(request).WithError(err)
	}

	request.Ctx = valuectx.WithValue(request.Ctx, "error", err)

	return handler(request)
}
