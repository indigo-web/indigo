package obtainer

import (
	"context"
	"github.com/fakefloordiv/indigo/http"
	"github.com/fakefloordiv/indigo/router/inbuilt/radix"
	routertypes "github.com/fakefloordiv/indigo/router/inbuilt/types"
	"github.com/fakefloordiv/indigo/types"
)

func DynamicObtainer(routes routertypes.RoutesMap) Obtainer {
	tree := getTree(routes)

	return func(ctx context.Context, req *types.Request) (context.Context, routertypes.HandlerFunc, error) {
		var methods routertypes.MethodsMap
		ctx, methods = tree.Match(ctx, req.Path)
		if methods == nil {
			return ctx, nil, http.ErrNotFound
		}

		handler := getHandler(req.Method, methods)
		if handler == nil {
			return ctx, nil, http.ErrMethodNotAllowed
		}

		return ctx, handler, nil
	}
}

func getTree(routes routertypes.RoutesMap) radix.Tree {
	tree := radix.NewTree()

	for k, v := range routes {
		tree.MustInsert(radix.MustParse(k), v)
	}

	return tree
}
