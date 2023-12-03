package inbuilt

import (
	"github.com/indigo-web/indigo/http"
	"github.com/indigo-web/indigo/internal/pathlib"
)

// Static adds a catcher of prefix, that automatically returns files from defined root
// directory
func (r *Router) Static(prefix, root string, mwares ...Middleware) *Router {
	pathReplacer := pathlib.NewPath(prefix, root)

	return r.Catch(prefix, func(request *http.Request) *http.Response {
		pathReplacer.Set(request.Path)

		return request.Respond().File(pathReplacer.Relative())
	}, mwares...)
}
