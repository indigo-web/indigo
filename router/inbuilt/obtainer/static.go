package obtainer

import (
	methods "github.com/indigo-web/indigo/http/method"
	"strings"

	"github.com/indigo-web/indigo/http"
	"github.com/indigo-web/indigo/http/status"
	"github.com/indigo-web/indigo/internal/functools"
	"github.com/indigo-web/indigo/internal/mapconv"
	"github.com/indigo-web/indigo/router/inbuilt/types"
	"github.com/indigo-web/indigo/valuectx"
)

func StaticObtainer(routes types.RoutesMap) Obtainer {
	allowedMethods := getAllowedMethodsMap(routes)

	return func(req *http.Request) (types.HandlerFunc, error) {
		methodsMap, found := routes[stripTrailingSlash(req.Path)]
		if !found {
			return nil, status.ErrNotFound
		}

		if handler := getHandler(req.Method, methodsMap); handler != nil {
			return handler, nil
		}

		req.Ctx = valuectx.WithValue(req.Ctx, "allow", allowedMethods[req.Path])

		return nil, status.ErrMethodNotAllowed
	}
}

func getAllowedMethodsMap(routes types.RoutesMap) map[string]string {
	allowedMethods := make(map[string]string, len(routes))

	for resource, methodsMap := range routes {
		keys := functools.Map(methods.ToString, mapconv.Keys(methodsMap))
		allowedMethods[resource] = strings.Join(keys, ",")
	}

	return allowedMethods
}

func getHandler(method methods.Method, methodsMap types.MethodsMap) types.HandlerFunc {
	handler, found := methodsMap[method]
	if !found {
		if method == methods.HEAD {
			return getHandler(methods.GET, methodsMap)
		}

		return nil
	}

	return handler.Fun
}
