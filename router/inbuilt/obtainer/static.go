package obtainer

import (
	"strings"

	"github.com/indigo-web/indigo/http"
	"github.com/indigo-web/indigo/http/method"
	"github.com/indigo-web/indigo/http/status"
	"github.com/indigo-web/indigo/router/inbuilt/types"
	"github.com/indigo-web/indigo/valuectx"
	"github.com/indigo-web/utils/ft"
	"github.com/indigo-web/utils/mapconv"
)

func StaticObtainer(routes types.RoutesMap) Obtainer {
	allowedMethods := getAllowedMethodsMap(routes)

	return func(req *http.Request) (types.HandlerFunc, error) {
		methodsMap, found := routes[stripTrailingSlash(req.Path.String)]
		if !found {
			return nil, status.ErrNotFound
		}

		if handler := getHandler(req.Method, methodsMap); handler != nil {
			return handler, nil
		}

		req.Ctx = valuectx.WithValue(req.Ctx, "allow", allowedMethods[req.Path.String])

		return nil, status.ErrMethodNotAllowed
	}
}

func getAllowedMethodsMap(routes types.RoutesMap) map[string]string {
	allowedMethods := make(map[string]string, len(routes))

	for resource, methodsMap := range routes {
		keys := ft.Map(method.ToString, mapconv.Keys(methodsMap))
		allowedMethods[resource] = strings.Join(keys, ",")
	}

	return allowedMethods
}

func getHandler(reqMethod method.Method, methodsMap types.MethodsMap) types.HandlerFunc {
	handler, found := methodsMap[reqMethod]
	if !found {
		if reqMethod == method.HEAD {
			return getHandler(method.GET, methodsMap)
		}

		return nil
	}

	return handler.Fun
}
