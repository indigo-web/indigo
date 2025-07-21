package inbuilt

import (
	"fmt"
	"github.com/indigo-web/indigo/http"
	"github.com/indigo-web/indigo/http/mime"
	"github.com/indigo-web/indigo/http/status"
	"os"
	"strings"
)

// Static adds a catcher of prefix, that automatically returns files from defined root
// directory
func (r *Router) Static(prefix, root string, mwares ...Middleware) *Router {
	stat, err := os.Stat(root)
	if err != nil {
		panic(err)
	}

	if !stat.IsDir() {
		panic(fmt.Sprintf("%s: not a directory", root))
	}

	fs := os.DirFS(root)

	return r.Get(prefix+"/:path...", func(request *http.Request) *http.Response {
		path := request.Vars.Value("path")
		if !isSafe(path) {
			return http.Error(request, status.ErrBadRequest)
		}

		file, err := fs.Open(path)
		if err != nil {
			return http.
				Error(request, status.ErrNotFound).
				String(err.(*os.PathError).Err.Error())
		}

		fstat, err := file.Stat()
		if err != nil {
			return http.
				Error(request, status.ErrInternalServerError).
				String(err.(*os.PathError).Err.Error())
		}

		return http.
			SizedStream(request, file, fstat.Size()).
			ContentType(mime.Guess(path))
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

		path = path[dot+1:]
	}

	return true
}
