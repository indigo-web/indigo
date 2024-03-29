package router

import (
	"github.com/indigo-web/indigo/http"
)

// Router is a general interface for any router compatible with indigo
// OnRequest called every time headers are parsed and ready to be processed
// OnError called once, and if it called, it means that connection will be
// closed anyway. So you can process the error, send some response,
// and when you are ready, just notify core that he can safely close
// the connection (even if it's already closed from client side).
type Router interface {
	OnStart() error
	OnRequest(request *http.Request) *http.Response
	OnError(request *http.Request, err error) *http.Response
}
