package mutator

import (
	"github.com/indigo-web/indigo/http"
	"github.com/indigo-web/indigo/router/inbuilt/internal/types"
)

func Alias(from, to string) types.Mutator {
	return func(request *http.Request) {
		if request.Path == from {
			request.Path = to
			request.Env.AliasFrom = from
		}
	}
}
