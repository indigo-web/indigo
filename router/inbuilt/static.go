package inbuilt

import (
	"github.com/indigo-web/indigo/http"
	"github.com/indigo-web/indigo/router/inbuilt/types"
	"os"
	"strings"
)

// Static adds a catcher of prefix, that automatically returns files from defined root
// directory
func (r *Router) Static(prefix, root string, mwares ...types.Middleware) *Router {
	prefix = stripTrailingSlash(prefix)
	root = mustTrailingSlash(root)

	return r.Catch(prefix, func(request *http.Request) *http.Response {
		path := root + stripLeadingSlash(strings.TrimPrefix(request.Path, prefix))
		stat, err := os.Stat(path)
		if err != nil {
			return request.Respond().WithError(err)
		}

		if stat.IsDir() {
			path += "/index.html"
		}

		return request.Respond().WithFile(path)
	}, mwares...)
}

func mustTrailingSlash(path string) string {
	return stripTrailingSlash(path) + "/"
}

func stripLeadingSlash(path string) string {
	for i := 0; i < len(path); i++ {
		if path[i] != '/' {
			return path[i:]
		}
	}

	return ""
}
