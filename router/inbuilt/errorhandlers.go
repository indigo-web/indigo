package inbuilt

import (
	"github.com/indigo-web/indigo/http"
	"github.com/indigo-web/indigo/http/method"
	"github.com/indigo-web/indigo/http/status"
)

func newErrorHandlers() errorHandlers {
	return errorHandlers{
		AllErrors:               genericErrorHandler,
		status.MethodNotAllowed: generic405Handler,
	}
}

func genericErrorHandler(request *http.Request) *http.Response {
	return http.Error(request, request.Env.Error)
}

func generic405Handler(request *http.Request) *http.Response {
	resp := request.Respond().Header("Allow", request.Env.AllowedMethods)

	if request.Method != method.OPTIONS {
		resp.Code(status.MethodNotAllowed)
	}

	return resp
}
