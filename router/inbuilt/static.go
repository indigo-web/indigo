package inbuilt

import (
	"github.com/indigo-web/indigo/http"
	"github.com/indigo-web/indigo/http/status"
	"github.com/indigo-web/indigo/internal/pathlib"
	"strings"
)

// Static adds a catcher of prefix, that automatically returns files from defined root
// directory
func (r *Router) Static(prefix, root string, mwares ...Middleware) *Router {
	pathReplacer := pathlib.NewPath(prefix, root)

	return r.Catch(prefix, func(request *http.Request) *http.Response {
		pathReplacer.Set(request.Path)
		path := pathReplacer.Relative()
		if !isSafe(path) {
			return http.Error(request, status.ErrNotFound)
		}

		return request.Respond().File(path)
	}, mwares...)
}

// isSafe checks for path traversal (basically - double dots)
func isSafe(path string) bool {
	for len(path) > 0 {
		dot := strings.IndexByte(path, '.')
		if dot == -1 {
			return true
		}

		if dot < len(path)-1 && path[dot+1] == '.' {
			return false
		}
	}

	return true
}
