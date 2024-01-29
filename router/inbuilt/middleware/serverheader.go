package middleware

import (
	"github.com/indigo-web/indigo/http"
	"github.com/indigo-web/indigo/router/inbuilt"
	"strings"
)

const DefaultServerHeader = "indigo"

func ServerHeader(customHeaders ...string) inbuilt.Middleware {
	value := strings.Join(customHeaders, " ")
	if len(value) == 0 {
		value = DefaultServerHeader
	}

	return func(next inbuilt.Handler, request *http.Request) *http.Response {
		return next(request).
			Header("Server", value)
	}
}
