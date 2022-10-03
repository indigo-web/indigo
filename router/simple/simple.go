package simple

import (
	"github.com/fakefloordiv/indigo/http/encodings"
	router2 "github.com/fakefloordiv/indigo/router"
	"github.com/fakefloordiv/indigo/types"
)

type Handler func(*types.Request) types.Response

// router TODO: add error handler. Simple does not mean castrated
type router struct {
	handler Handler
}

func NewRouter(handler Handler) router2.Router {
	return router{
		handler: handler,
	}
}

func (r router) OnRequest(request *types.Request) types.Response {
	return r.handler(request)
}

func (router) OnError(_ *types.Request, err error) types.Response {
	return types.WithError(err)
}

func (router) GetContentEncodings() encodings.ContentEncodings {
	return encodings.NewContentEncodings()
}
