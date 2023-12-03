package inbuilt

import (
	"github.com/indigo-web/indigo/http/method"
	"github.com/indigo-web/indigo/http/status"
	"github.com/indigo-web/indigo/router/inbuilt/internal/types"
	"strings"
)

type Handler = types.Handler

type errorHandlers map[status.Code]Handler

type routesMapEntry struct {
	methodsMap types.MethodsMap
	allow      string
}

type routesMap map[string]routesMapEntry

func (r routesMap) Add(path string, m method.Method, handler Handler) {
	entry := r[path]
	entry.methodsMap[m] = handler
	entry.allow = getAllowString(entry.methodsMap)
	r[path] = entry
}

func getAllowString(methodsMap types.MethodsMap) (allowed string) {
	for i, handler := range methodsMap {
		if handler == nil {
			continue
		}

		allowed += method.Method(i).String() + ","
	}

	return strings.TrimRight(allowed, ",")
}
