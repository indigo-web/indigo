package router

import (
	"fmt"
	"indigo/http"
	"indigo/internal"
	"indigo/types"
)

const DefaultResponseBufferSize = 1024

type (
	Handler  func(request *types.Request) *types.ResponseStruct
	Handlers map[string]Handler
)

type DefaultRouter struct {
	handlers Handlers
	respBuff []byte
}

func NewDefaultRouter() *DefaultRouter {
	return &DefaultRouter{
		handlers: make(Handlers, 5),
		respBuff: make([]byte, 0, DefaultResponseBufferSize),
	}
}

func (d *DefaultRouter) Route(path string, handler Handler) {
	d.handlers[path] = handler
}

func (d DefaultRouter) OnRequest(req *types.Request, writeResponse types.ResponseWriter) error {
	handler, found := d.handlers[internal.B2S(req.Path)]

	var resp *types.ResponseStruct

	if !found {
		resp = types.Response().
			WithCode(http.StatusNotFound).
			WithBodyString("404 requested page not found")
	} else {
		resp = handler(req)
	}

	return writeResponse(http.RenderHTTPResponse(
		d.respBuff[:0],
		req.Protocol.Raw(),
		http.ByteStatusCodes[resp.Code],
		http.GetStatus(resp.Code),
		resp.GetHeaders(),
		resp.Body,
	))
}

func (d DefaultRouter) OnError(err error) {
	fmt.Println("fatal: got err:", err)
}
