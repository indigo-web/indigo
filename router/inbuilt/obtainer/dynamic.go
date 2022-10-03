package obtainer

import (
	"context"
	"strings"

	"github.com/fakefloordiv/indigo/http"
	methods "github.com/fakefloordiv/indigo/http/method"
	"github.com/fakefloordiv/indigo/internal/functools"
	"github.com/fakefloordiv/indigo/internal/mapconv"
	"github.com/fakefloordiv/indigo/router/inbuilt/radix"
	routertypes "github.com/fakefloordiv/indigo/router/inbuilt/types"
	"github.com/fakefloordiv/indigo/types"
	"github.com/fakefloordiv/indigo/valuectx"
)

func DynamicObtainer(routes routertypes.RoutesMap) Obtainer {
	tree := getTree(routes)

	return func(ctx context.Context, req *types.Request) (context.Context, routertypes.HandlerFunc, error) {
		var payload *radix.Payload
		ctx, payload = tree.Match(ctx, req.Path)
		if payload == nil {
			return ctx, nil, http.ErrNotFound
		}

		handler := getHandler(req.Method, payload.MethodsMap)
		if handler == nil {
			ctx = valuectx.WithValue(ctx, "allow", payload.Allow)

			return ctx, nil, http.ErrMethodNotAllowed
		}

		return ctx, handler, nil
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
