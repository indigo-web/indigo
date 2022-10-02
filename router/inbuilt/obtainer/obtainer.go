package obtainer

import (
	"context"
	"github.com/fakefloordiv/indigo/router/inbuilt/radix"
	routertypes "github.com/fakefloordiv/indigo/router/inbuilt/types"
	"github.com/fakefloordiv/indigo/types"
)

type Obtainer func(context.Context, *types.Request) (context.Context, routertypes.HandlerFunc, error)

// Auto simply checks all the routes, and returns a corresponding obtainer.
// In case all the routes are static, StaticObtainer is returned.
// In case at least one route is dynamic, DynamicObtainer is returned
func Auto(routes routertypes.RoutesMap) Obtainer {
	for k := range routes {
		if !radix.MustParse(k).IsStatic() {
			return DynamicObtainer(routes)
		}
	}

	return StaticObtainer(routes)
}
