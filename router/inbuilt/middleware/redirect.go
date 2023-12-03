package middleware

import (
	"github.com/indigo-web/indigo/http"
	"github.com/indigo-web/indigo/http/status"
	"github.com/indigo-web/indigo/router/inbuilt"
)

func Redirect(from, to string) inbuilt.Middleware {
	return func(next inbuilt.Handler, request *http.Request) *http.Response {
		if request.Path != from {
			return next(request)
		}

		return request.Respond().
			Code(status.TemporaryRedirect).
			Header("Location", to)
	}
}
