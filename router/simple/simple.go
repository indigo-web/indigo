package simple

import (
	"github.com/indigo-web/indigo/http"
	"github.com/indigo-web/indigo/router"
)

var _ router.Router = Router{}

type (
	Handler      func(*http.Request) *http.Response
	ErrorHandler func(*http.Request, error) *http.Response
)

type Router struct {
	handler    Handler
	errHandler ErrorHandler
}

func NewRouter(handler Handler, errHandler ErrorHandler) Router {
	return Router{
		handler:    handler,
		errHandler: errHandler,
	}
}

func (r Router) OnStart() error {
	return nil
}

func (r Router) OnRequest(request *http.Request) *http.Response {
	return r.handler(request)
}

func (r Router) OnError(request *http.Request, err error) *http.Response {
	return r.errHandler(request, err)
}
