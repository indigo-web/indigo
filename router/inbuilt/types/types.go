package types

import (
	"github.com/indigo-web/indigo/http"
	"github.com/indigo-web/indigo/http/method"
	"github.com/indigo-web/indigo/http/status"
	"strings"
)

type (
	Handler func(*http.Request) http.Response
	// Middleware works like a chain of nested calls, next may be even directly
	// handler. But if we are not a closing middleware, we will call next
	// middleware that is simply a partial middleware with already provided next
	Middleware func(next Handler, request *http.Request) http.Response
	MethodsMap [method.Count]Handler
)

type ErrHandlers struct {
	handlers  map[status.Code]Handler
	universal Handler
}

func NewErrHandlers() *ErrHandlers {
	return &ErrHandlers{
		handlers: make(map[status.Code]Handler),
	}
}

func (e *ErrHandlers) Set(code status.Code, handler Handler) {
	e.handlers[code] = handler
}

func (e *ErrHandlers) SetUniversal(universal Handler) {
	e.universal = universal
}

func (e *ErrHandlers) Get(err status.HTTPError) Handler {
	handler := e.handlers[err.Code]
	if handler == nil {
		handler = e.universal
	}

	return handler
}

type routesMapEntry struct {
	methodsMap MethodsMap
	allow      string
}

type RoutesMap map[string]routesMapEntry

func (r RoutesMap) Get(path string) (MethodsMap, string, bool) {
	entry, found := r[path]

	return entry.methodsMap, entry.allow, found
}

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
