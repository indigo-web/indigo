package types

import (
	"github.com/indigo-web/indigo/http"
	"github.com/indigo-web/indigo/http/method"
)

type (
	// Handler is a function for processing a request. Using named return as
	// underscore just in order to be able to make an empty return
	Handler    func(*http.Request) (_ *http.Response)
	MethodsMap [method.Count]Handler
	Mutator    func(request *http.Request)
)
