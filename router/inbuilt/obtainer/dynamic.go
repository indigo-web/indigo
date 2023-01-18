package obtainer

import (
	"strings"

	"github.com/fakefloordiv/indigo/http/status"

	"github.com/fakefloordiv/indigo/http"
	methods "github.com/fakefloordiv/indigo/http/method"
	"github.com/fakefloordiv/indigo/internal/functools"
	"github.com/fakefloordiv/indigo/internal/mapconv"
	"github.com/fakefloordiv/indigo/router/inbuilt/radix"
	routertypes "github.com/fakefloordiv/indigo/router/inbuilt/types"
	"github.com/fakefloordiv/indigo/valuectx"
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
