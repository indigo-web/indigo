package mutator

import (
	"github.com/indigo-web/indigo/http"
	"github.com/indigo-web/indigo/http/method"
	"github.com/indigo-web/indigo/router/inbuilt/internal"
)

func getAllowedMethods(methods []method.Method) (lut [method.Count + 1]bool) {
	if len(methods) == 0 {
		return getAllowedMethods(method.List)
	}

	for _, m := range methods {
		lut[m] = true
	}

	return lut
}

func Alias(from, to string, forMethods ...method.Method) internal.Mutator {
	allowedMethods := getAllowedMethods(forMethods)

	return func(request *http.Request) {
		if allowedMethods[request.Method] && request.Path == from {
			request.Path = to
			request.Env.AliasFrom = from
		}
	}
}
