package simple

import (
	"github.com/indigo-web/indigo/http"
	"github.com/indigo-web/indigo/http/encodings"
	router2 "github.com/indigo-web/indigo/router"
)

type (
	Handler      func(*http.Request) http.Response
	ErrorHandler func(*http.Request, error) http.Response
)

type router struct {
	handler    Handler
	errHandler ErrorHandler
}

func NewRouter(handler Handler, errHandler ErrorHandler) router2.Router {
	return router{
		handler:    handler,
		errHandler: errHandler,
	}
}

func (r router) OnRequest(request *http.Request) http.Response {
	return r.handler(request)
}

func (r router) OnError(request *http.Request, err error) http.Response {
	return r.errHandler(request, err)
}

func (router) GetContentEncodings() encodings.Decoders {
	return encodings.NewContentDecoders()
}
