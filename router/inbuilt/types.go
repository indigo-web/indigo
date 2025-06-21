package inbuilt

import (
	"github.com/indigo-web/indigo/http"
	"github.com/indigo-web/indigo/http/method"
	"github.com/indigo-web/indigo/http/status"
	"github.com/indigo-web/indigo/router/inbuilt/internal/radix"
	"strings"
)

type (
	Handler       func(*http.Request) *http.Response
	errorHandlers map[status.Code]Handler
	methodLUT     [method.Count + 1]Handler
)

type endpoint struct {
	methods methodLUT
	allow   string
}

type (
	routesMap map[string]endpoint
	radixTree = *radix.Node[endpoint]
)

func (r routesMap) Add(path string, m method.Method, handler Handler) {
	entry := r[path]
	entry.methods[m] = handler
	entry.allow = getAllowString(entry.methods)
	r[path] = entry
}

func getAllowString(methods methodLUT) (allowed string) {
	for i, handler := range methods {
		if handler == nil {
			continue
		}

		allowed += method.Method(i).String() + ","
	}

	return strings.TrimRight(allowed, ",")
}
