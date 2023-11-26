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
		file, err := request.Respond().WithFile(root + path)
		if err != nil {
			return request.Respond().WithError(err)
		}

		return file
	}, mwares...)
}

func mustTrailingSlash(path string) string {
	return stripTrailingSlash(path) + "/"
}
