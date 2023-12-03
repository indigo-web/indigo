package middleware

import (
	"context"
	"github.com/indigo-web/indigo/http"
	"github.com/indigo-web/indigo/router/inbuilt"
)

func CustomContext(ctx context.Context) inbuilt.Middleware {
	return func(next inbuilt.Handler, request *http.Request) *http.Response {
		request.Ctx = ctx

		return next(request)
	}
}
