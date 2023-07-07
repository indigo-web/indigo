package types

import (
	"github.com/indigo-web/indigo/http"

	"github.com/indigo-web/indigo/http/method"
)

type (
	HandlerFunc func(*http.Request) http.Response
	ErrHandlers map[error]HandlerFunc
	// Middleware works like a chain of nested calls, next may be even directly
	// handler. But if we are not a closing middleware, we will call next
	// middleware that is simply a partial middleware with already provided next
	Middleware func(next HandlerFunc, request *http.Request) http.Response
)

type (
	MethodsMap [method.Count]*HandlerObject
	RoutesMap  map[string]MethodsMap

	HandlerObject struct {
		Fun         HandlerFunc
		Middlewares []Middleware
	}
)
