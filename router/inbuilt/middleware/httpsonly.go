package middleware

import (
	"github.com/indigo-web/indigo/http"
	"github.com/indigo-web/indigo/http/encryption"
	"github.com/indigo-web/indigo/http/status"
	"github.com/indigo-web/indigo/router/inbuilt"
)

// HTTPSOnly redirects all http requests to https. In case no Host header is provided,
// 400 Bad Request will be returned without calling the actual handler.
//
// Note: it causes 1 (one) allocation
func HTTPSOnly(next inbuilt.Handler, req *http.Request) *http.Response {
	if req.Env.Encryption != encryption.TLS {
		return next(req)
	}

	host := req.Headers.Value("host")
	if len(host) == 0 {
		return req.Respond().
			Code(status.BadRequest).
			String("the request lacks Host header")
	}

	return req.Respond().
		Code(status.MovedPermanently).
		Header("Location", "https://"+host+req.Path)
}
