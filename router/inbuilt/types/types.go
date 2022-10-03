package types

import (
	"context"

	methods "github.com/fakefloordiv/indigo/http/method"
	"github.com/fakefloordiv/indigo/types"
)

type (
	HandlerFunc func(context.Context, *types.Request) types.Response
	// Middleware works like a chain of nested calls, next may be even directly
	// handler. But if we are not a closing middleware, we will call next
	// middleware that is simply a partial middleware with already provided next
	Middleware func(ctx context.Context, next HandlerFunc, request *types.Request) types.Response
)

type (
	MethodsMap map[methods.Method]*HandlerObject
	RoutesMap  map[string]MethodsMap

	HandlerObject struct {
		Fun         HandlerFunc
		Middlewares []Middleware
	}
)
