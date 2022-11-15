package inbuilt

import (
	"github.com/fakefloordiv/indigo/http/status"
	routertypes "github.com/fakefloordiv/indigo/router/inbuilt/types"

	"github.com/fakefloordiv/indigo/http"
)

/*
This file is responsible for default responses for errors. Currently, this
functionality is not needed because response builder already has WithError()
method that automatically determines error and returns a corresponding code
and body (as a text of the error, just like http wants). The only exception
is 405 Method Not Allowed, because http requires Allow header, that is why
we have to customize a behaviour of response with such a code
*/

func newErrorHandlers() routertypes.ErrHandlers {
	return routertypes.ErrHandlers{
		status.ErrMethodNotAllowed: defaultMethodNotAllowedHandler,
	}
}

func defaultMethodNotAllowedHandler(request *http.Request) http.Response {
	response := request.Respond.WithError(status.ErrMethodNotAllowed)

	if allow, ok := request.Ctx.Value("allow").(string); ok {
		response = response.WithHeader("Allow", allow)
	}

	return response
}
