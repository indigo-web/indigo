package simple

import (
	"github.com/indigo-web/indigo/http"
	"github.com/indigo-web/indigo/router"
)

var _ router.Fabric = new(Router)

type Handler func(*http.Request) *http.Response

type Router struct {
	handler, errHandler Handler
}

func New(handler, errHandler Handler) *Router {
	return &Router{
		handler:    handler,
		errHandler: errHandler,
	}
}

func (r Router) Initialize() router.Router {
	return r
}

func (r Router) OnRequest(request *http.Request) *http.Response {
	return r.handler(request)
}

func (r Router) OnError(request *http.Request, err error) *http.Response {
	request.Env.Error = err

	return r.errHandler(request)
}
