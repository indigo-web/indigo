package inbuilt

import (
	"github.com/indigo-web/indigo/http"
	"github.com/indigo-web/indigo/router/inbuilt/types"
	"strings"
)

// Static adds a catcher of prefix, that automatically returns files from defined root
// directory
func (r *Router) Static(prefix, root string, mwares ...types.Middleware) *Router {
	prefix = mustTrailingSlash(prefix)
	root = mustTrailingSlash(root)

	return r.Catch(prefix, func(request *http.Request) *http.Response {
		path := strings.TrimPrefix(request.Path, prefix)

		return request.Respond().WithFile(root + path)
	}, mwares...)
}

func mustTrailingSlash(path string) string {
	return stripTrailingSlash(path) + "/"
}
