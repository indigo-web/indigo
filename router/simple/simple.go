package simple

import (
	"context"

	"github.com/fakefloordiv/indigo/http/encodings"
	"github.com/fakefloordiv/indigo/http/status"
	router2 "github.com/fakefloordiv/indigo/router"
	"github.com/fakefloordiv/indigo/router/inbuilt"
	"github.com/fakefloordiv/indigo/types"
)

var defaultErrResponse = types.WithResponse.
	WithCode(status.BadRequest).
	WithBody(`<h1 align="center">400 Bad Request</h1>`)

// router TODO: add error handler and content encodings setter. Simple does not mean castrated
type router struct {
	handler inbuilt.HandlerFunc
}

func NewRouter(handler inbuilt.HandlerFunc) router2.Router {
	return router{
		handler: handler,
	}
}

func (r router) OnRequest(request *types.Request, render types.Render) error {
	return render(r.handler(context.Background(), request))
}

func (router) OnError(_ *types.Request, render types.Render, _ error) {
	_ = render(defaultErrResponse)
}

func (router) GetContentEncodings() encodings.ContentEncodings {
	return encodings.NewContentEncodings()
}
