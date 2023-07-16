package obtainer

import (
	"github.com/indigo-web/indigo/http"
	"github.com/indigo-web/indigo/http/method"
	"github.com/indigo-web/indigo/http/status"
	"github.com/indigo-web/indigo/router/inbuilt/types"
	"github.com/indigo-web/indigo/valuectx"
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

	for resource, methods := range routes {
		allowedMethods[resource] = getAllowedMethodsString(methods)
	}

	return allowedMethods
}

func getAllowedMethodsString(methods types.MethodsMap) (str string) {
	for i := range methods {
		if methods[i] == nil {
			continue
		}

		str += method.ToString(method.Method(i)) + ","
	}

	if len(str) > 0 {
		// remove trailing comma
		str = str[:len(str)-1]
	}

	return str
}

func getHandler(reqMethod method.Method, methodsMap types.MethodsMap) types.HandlerFunc {
	handler := methodsMap[reqMethod]
	if handler == nil {
		if reqMethod == method.HEAD {
			return getHandler(method.GET, methodsMap)
		}

		return nil
	}

	return handler.Fun
}
