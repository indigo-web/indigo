package types

import (
	"github.com/indigo-web/indigo/http"
	"github.com/indigo-web/indigo/http/method"
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
