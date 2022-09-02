package simple

import (
	"github.com/fakefloordiv/indigo/http/render"
	"github.com/fakefloordiv/indigo/http/status"
	router2 "github.com/fakefloordiv/indigo/router"
	"github.com/fakefloordiv/indigo/router/inbuilt"
	"github.com/fakefloordiv/indigo/types"
)

var defaultErrResponse = types.WithResponse.
	WithCode(status.BadRequest).
	WithBody(`<h1 align="center">400 Bad Request</h1>`)

type router struct {
	handler  inbuilt.HandlerFunc
	renderer *render.Renderer
}

func NewRouter(handler inbuilt.HandlerFunc) router2.Router {
	return router{
		handler:  handler,
		renderer: render.NewRenderer(nil),
	}
}

func (router) OnStart() {}

func (r router) OnRequest(request *types.Request, respWriter types.ResponseWriter) error {
	return r.renderer.Response(request.Proto, r.handler(request), respWriter)
}

func (r router) OnError(request *types.Request, respWriter types.ResponseWriter, _ error) {
	_ = r.renderer.Response(request.Proto, defaultErrResponse, respWriter)
}
