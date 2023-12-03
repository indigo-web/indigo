package types

import (
	"github.com/indigo-web/indigo/http"
	"github.com/indigo-web/indigo/http/method"
)

type (
	Handler    func(*http.Request) *http.Response
	MethodsMap [method.Count]Handler
)
