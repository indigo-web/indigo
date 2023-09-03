package types

import (
	"github.com/indigo-web/indigo/http"
	"github.com/indigo-web/indigo/http/method"
	"strings"
)

type (
	Handler     func(*http.Request) http.Response
	ErrHandlers map[error]Handler
	// Middleware works like a chain of nested calls, next may be even directly
	// handler. But if we are not a closing middleware, we will call next
	// middleware that is simply a partial middleware with already provided next
	Middleware func(next Handler, request *http.Request) http.Response
	MethodsMap [method.Count]Handler
)

type routesMapEntry struct {
	methodsMap MethodsMap
	allow      string
}

type RoutesMap struct {
	m map[string]routesMapEntry
}

func NewRoutesMap() RoutesMap {
	return RoutesMap{
		m: make(map[string]routesMapEntry),
	}
}

func (r RoutesMap) Get(path string) (MethodsMap, string, bool) {
	entry, found := r.m[path]

	return entry.methodsMap, entry.allow, found
}

func (r RoutesMap) Add(path string, m method.Method, handler Handler) {
	entry := r.m[path]
	entry.methodsMap[m] = handler
	entry.allow = getAllowString(entry.methodsMap)
	r.m[path] = entry
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
