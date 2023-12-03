package inbuilt

import (
	"github.com/indigo-web/indigo/http"
	"github.com/indigo-web/indigo/http/status"
)

func newErrorHandlers() errorHandlers {
	return errorHandlers{
		AllErrors:               defaultAllErrorsHandler,
		status.MethodNotAllowed: defaultMethodNotAllowedHandler,
	}
}

func defaultAllErrorsHandler(request *http.Request) *http.Response {
	return request.Respond().Error(request.Env.Error)
}

func defaultMethodNotAllowedHandler(request *http.Request) *http.Response {
	return request.Respond().
		Error(status.ErrMethodNotAllowed).
		Header("Allow", request.Env.AllowMethods)
}
