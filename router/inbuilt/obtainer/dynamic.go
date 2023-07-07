package obtainer

import (
	"github.com/indigo-web/indigo/http/status"

	"github.com/indigo-web/indigo/http"
	"github.com/indigo-web/indigo/router/inbuilt/radix"
	"github.com/indigo-web/indigo/router/inbuilt/types"
	"github.com/indigo-web/indigo/valuectx"
)

func DynamicObtainer(routes types.RoutesMap) Obtainer {
	tree := getTree(routes)

	return func(req *http.Request) (types.HandlerFunc, error) {
		payload := tree.Match(req.Path.Params, stripTrailingSlash(req.Path.String))
		if payload == nil {
			return nil, status.ErrNotFound
		}

		handler := getHandler(req.Method, payload.MethodsMap)
		if handler == nil {
			req.Ctx = valuectx.WithValue(req.Ctx, "allow", payload.Allow)

			return nil, status.ErrMethodNotAllowed
		}

		return handler, nil
	}
}

func getTree(routes types.RoutesMap) radix.Tree {
	tree := radix.NewTree()

	for resource, methods := range routes {
		tree.MustInsert(radix.MustParse(resource), radix.Payload{
			MethodsMap: methods,
			Allow:      getAllowedMethodsString(methods),
		})
	}

	return tree
}
