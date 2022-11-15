package obtainer

import (
	"github.com/fakefloordiv/indigo/http/status"
	"strings"

	"github.com/fakefloordiv/indigo/http"
	methods2 "github.com/fakefloordiv/indigo/http/method"
	"github.com/fakefloordiv/indigo/internal/functools"
	"github.com/fakefloordiv/indigo/internal/mapconv"
	routertypes "github.com/fakefloordiv/indigo/router/inbuilt/types"
	"github.com/fakefloordiv/indigo/valuectx"
)

func StaticObtainer(routes routertypes.RoutesMap) Obtainer {
	allowedMethods := getAllowedMethodsMap(routes)

	return func(req *http.Request) (routertypes.HandlerFunc, error) {
		methods, found := routes[req.Path]
		if !found {
			return nil, status.ErrNotFound
		}

		if handler := getHandler(req.Method, methods); handler != nil {
			return handler, nil
		}

		req.Ctx = valuectx.WithValue(req.Ctx, "allow", allowedMethods[req.Path])

		return nil, status.ErrMethodNotAllowed
	}
}

func getAllowedMethodsMap(routes routertypes.RoutesMap) map[string]string {
	allowedMethods := make(map[string]string, len(routes))

	for resource, methods := range routes {
		keys := functools.Map(methods2.ToString, mapconv.Keys(methods))
		allowedMethods[resource] = strings.Join(keys, ",")
	}

	return allowedMethods
}

func getHandler(method methods2.Method, methods routertypes.MethodsMap) routertypes.HandlerFunc {
	handler, found := methods[method]
	if !found {
		if method == methods2.HEAD {
			return getHandler(methods2.GET, methods)
		}

		return nil
	}

	return handler.Fun
}
