package inbuilt

import (
	"context"
	"github.com/fakefloordiv/indigo/http"

	"github.com/fakefloordiv/indigo/types"
)

/*
This file is responsible for default responses for errors. Currently, this
functionality is not needed because response builder already has WithError()
method that automatically determines error and returns a corresponding code
and body (as a text of the error, just like http wants). The only exception
is 405 Method Not Allowed, because http requires Allow header, that is why
we have to customize a behaviour of response with such a code
*/

func newErrorHandlers() errHandlers {
	return errHandlers{
		http.ErrMethodNotAllowed: defaultMethodNotAllowedHandler,
	}
}

func defaultMethodNotAllowedHandler(ctx context.Context, _ *types.Request) types.Response {
	allow := ctx.Value("allow").(string)

	return types.WithResponse.
		WithError(http.ErrMethodNotAllowed).
		WithHeader("Allow", allow)
}
