package middleware

import (
	"github.com/indigo-web/indigo/http"
	"github.com/indigo-web/indigo/router/inbuilt"
)

// Autocompress prepends automatic compression options to responses.
func Autocompress(next inbuilt.Handler, request *http.Request) *http.Response {
	return next(request).Compress()
}
