package simple

import (
	"github.com/fakefloordiv/indigo/http/encodings"
	router2 "github.com/fakefloordiv/indigo/router"
	"github.com/fakefloordiv/indigo/types"
)

type (
	Handler      func(*types.Request) types.Response
	ErrorHandler func(*types.Request, error) types.Response
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

func (r router) OnRequest(request *types.Request) types.Response {
	return r.handler(request)
}

func (r router) OnError(request *types.Request, err error) types.Response {
	return r.errHandler(request, err)
}

func (router) GetContentEncodings() encodings.Decoders {
	return encodings.NewContentDecoders()
}
