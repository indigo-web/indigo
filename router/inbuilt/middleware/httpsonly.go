package middleware

import (
	"github.com/indigo-web/indigo/http"
	"github.com/indigo-web/indigo/http/status"
	"github.com/indigo-web/indigo/router/inbuilt"
	"strings"
)

type HTTPOnlyParams struct {
	// RedirectTo defines the host, where the user will be redirected.
	// If empty, value from Host header will be used
	RedirectTo string
	// Port is added to the host value.
	// If empty, implicitly default 443 port will be used
	Port string
}

// HTTPSOnly redirects all http requests to https. In case no Host header is provided,
// 400 Bad Request will be returned without calling the actual handler.
func HTTPSOnly(optionalParams ...HTTPOnlyParams) inbuilt.Middleware {
	params := optional(optionalParams, HTTPOnlyParams{})

	return func(next inbuilt.Handler, request *http.Request) *http.Response {
		if request.Env.Encryption != 0 {
			return next(request)
		}

		host := params.RedirectTo

		if len(host) == 0 {
			host = removePort(request.Headers.Value("host"))
			if len(host) == 0 {
				return request.Respond().
					Code(status.BadRequest).
					String("no Host header")
			}
		}

		host += ":" + params.Port
		if len(request.Env.AliasFrom) > 0 {
			host += request.Env.AliasFrom
		} else {
			host += request.Path
		}

		return request.Respond().
			Code(status.MovedPermanently).
			Header("Location", "https://"+host)
	}
}

func removePort(str string) string {
	if colon := strings.IndexByte(str, ':'); colon != -1 {
		return str[:colon]
	}

	return str
}

func optional[T any](custom []T, default_ T) T {
	if len(custom) == 0 {
		return default_
	}

	return custom[0]
}
