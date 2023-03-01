package obtainer

import (
	"strings"

	"github.com/indigo-web/indigo/v2/http/status"

	"github.com/indigo-web/indigo/v2/http"
	methods "github.com/indigo-web/indigo/v2/http/method"
	"github.com/indigo-web/indigo/v2/internal/functools"
	"github.com/indigo-web/indigo/v2/internal/mapconv"
	"github.com/indigo-web/indigo/v2/router/inbuilt/radix"
	routertypes "github.com/indigo-web/indigo/v2/router/inbuilt/types"
	"github.com/indigo-web/indigo/v2/valuectx"
)

func DynamicObtainer(routes routertypes.RoutesMap) Obtainer {
	tree := getTree(routes)

	return func(req *http.Request) (routertypes.HandlerFunc, error) {
		var payload *radix.Payload
		req.Ctx, payload = tree.Match(req.Ctx, stripTrailingSlash(req.Path))
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

func getTree(routes routertypes.RoutesMap) radix.Tree {
	tree := radix.NewTree()

	for k, v := range routes {
		tree.MustInsert(radix.MustParse(k), radix.Payload{
			MethodsMap: v,
			Allow:      strings.Join(functools.Map(methods.ToString, mapconv.Keys(v)), ","),
		})
	}

	return tree
}
