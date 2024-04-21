package router

import (
	"github.com/indigo-web/indigo/http"
)

// Fabric constructs fully initialized routers
type Fabric interface {
	Initialize() Router
}

// Router is a completely initialized router, returned by Router
type Router interface {
	OnRequest(request *http.Request) *http.Response
	OnError(request *http.Request, err error) *http.Response
}
