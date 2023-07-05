package obtainer

import (
	"strings"

	"github.com/indigo-web/indigo/http/status"

	"github.com/indigo-web/indigo/http"
	"github.com/indigo-web/indigo/http/method"
	"github.com/indigo-web/indigo/router/inbuilt/radix"
	"github.com/indigo-web/indigo/router/inbuilt/types"
	"github.com/indigo-web/indigo/valuectx"
	"github.com/indigo-web/utils/ft"
	"github.com/indigo-web/utils/mapconv"
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

	for k, v := range routes {
		tree.MustInsert(radix.MustParse(k), radix.Payload{
			MethodsMap: v,
			Allow:      strings.Join(ft.Map(method.ToString, mapconv.Keys(v)), ","),
		})
	}

	return tree
}
