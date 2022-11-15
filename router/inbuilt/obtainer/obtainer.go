package obtainer

import (
	"github.com/fakefloordiv/indigo/http"

	"github.com/fakefloordiv/indigo/internal/mapconv"
	"github.com/fakefloordiv/indigo/router/inbuilt/radix"
	routertypes "github.com/fakefloordiv/indigo/router/inbuilt/types"
)

type (
	Obtainer    func(*http.Request) (routertypes.HandlerFunc, error)
	constructor func(routertypes.RoutesMap) Obtainer
)

// Auto simply checks all the routes, and returns a corresponding obtainer.
// In case all the routes are static, StaticObtainer is returned.
// In case at least one route is dynamic, DynamicObtainer is returned
func Auto(routes routertypes.RoutesMap) Obtainer {
	return getObtainer(mapconv.Keys(routes))(routes)
}

func getObtainer(routes []string) constructor {
	for _, route := range routes {
		if !radix.MustParse(route).IsStatic() {
			return DynamicObtainer
		}
	}

	return StaticObtainer
}
