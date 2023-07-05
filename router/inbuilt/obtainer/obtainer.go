package obtainer

import (
	"github.com/indigo-web/indigo/http"

	"github.com/indigo-web/indigo/router/inbuilt/radix"
	"github.com/indigo-web/indigo/router/inbuilt/types"
	"github.com/indigo-web/utils/mapconv"
)

type (
	Obtainer    func(*http.Request) (types.HandlerFunc, error)
	constructor func(types.RoutesMap) Obtainer
)

// Auto simply checks all the routes, and returns a corresponding obtainer.
// In case all the routes are static, StaticObtainer is returned.
// In case at least one route is dynamic, DynamicObtainer is returned
func Auto(routes types.RoutesMap) Obtainer {
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

// stripTrailingSlash just removes a trailing slash of request path in case it is presented.
// Note: this removes only one trailing slash. In case 2 or more are presented they'll be treated
// as an ordinary part of the path so won't be stripped
func stripTrailingSlash(path string) string {
	if path[len(path)-1] == '/' && len(path) > 1 {
		return path[:len(path)-1]
	}

	return path
}
