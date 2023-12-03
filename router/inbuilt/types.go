package inbuilt

import (
	"github.com/indigo-web/indigo/http"
	"github.com/indigo-web/indigo/http/method"
	"github.com/indigo-web/indigo/http/status"
	"github.com/indigo-web/indigo/router/inbuilt/internal/types"
	"strings"
)

type (
	Handler = types.Handler
	// Middleware works like a chain of nested calls, next may be even directly
	// handler. But if we are not a closing middleware, we will call next
	// middleware that is simply a partial middleware with already provided next
	Middleware func(next Handler, request *http.Request) *http.Response
	Catcher    struct {
		Prefix  string
		Handler Handler
	}
	MethodsMap = types.MethodsMap
)

type errorHandlers map[status.Code]Handler

type routesMapEntry struct {
	methodsMap MethodsMap
	allow      string
}

type RoutesMap map[string]routesMapEntry

func (r RoutesMap) Add(path string, m method.Method, handler Handler) {
	entry := r[path]
	entry.methodsMap[m] = handler
	entry.allow = getAllowString(entry.methodsMap)
	r[path] = entry
}

func getAllowString(methodsMap MethodsMap) (allowed string) {
	for i, handler := range methodsMap {
		if handler == nil {
			continue
		}

		allowed += method.Method(i).String() + ","
	}

	return strings.TrimRight(allowed, ",")
}
