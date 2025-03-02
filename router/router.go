package router

import (
	"github.com/indigo-web/indigo/http"
)

// Builder is used to type-safely separate router into two distinct stages: initialization
// (e.g. registering all the endpoints) and compilation (building the actual router). Builder
// itself represents the first stage.
type Builder interface {
	Build() Router
}

// Router is the second stage of a router, usually produced via Builder.Build. It is normally
// used internally in runtime by HTTP core.
type Router interface {
	OnRequest(request *http.Request) *http.Response
	OnError(request *http.Request, err error) *http.Response
}
