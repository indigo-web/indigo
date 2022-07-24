package router

import (
	"fmt"
	"indigo/http"
	"indigo/internal"
	"indigo/types"
	"strconv"
)

const DefaultResponseBufferSize = 1024

type (
	Handler  func(request *types.Request) types.Response
	Handlers map[string]Handler
)

type DefaultRouter struct {
	handlers Handlers
	respBuff []byte
}

func NewDefaultRouter() *DefaultRouter {
	respBuffWithProto := make([]byte, len(http.BytesHTTP11), DefaultResponseBufferSize)
	copy(respBuffWithProto, http.BytesHTTP11)

	return &DefaultRouter{
		handlers: make(Handlers, 5),
		respBuff: append(respBuffWithProto, ' '),
	}
}

func (d *DefaultRouter) Route(path string, handler Handler) {
	d.handlers[path] = handler
}

func (d DefaultRouter) OnRequest(req *types.Request, writeResponse types.ResponseWriter) (err error) {
	defer req.Reset()
	handler, found := d.handlers[internal.B2S(req.Path)]

	var resp types.Response

	if !found {
		resp = types.Response{
			Code: http.StatusNotFound,
			Body: []byte("404 requested page not found"),
		}
	} else {
		resp = handler(req)
	}

	resp = prepareResponse(resp)
	err, d.respBuff = types.WriteResponse(d.respBuff[:len(http.BytesHTTP11)+1], resp, writeResponse)

	return err
}

func (d DefaultRouter) OnError(err error) {
	fmt.Println("fatal: got err:", err)
}

func prepareResponse(response types.Response) types.Response {
	if response.Code == 0 {
		response.Code = http.StatusOk
	}

	bodyLen := strconv.Itoa(len(response.Body))

	if response.Headers == nil {
		response.Headers = [][]byte{
			[]byte("Server: indigo"),
			[]byte("Connection: keep-alive"),
			[]byte("Content-Length: " + bodyLen),
		}
	} else {
		response.Headers = append(response.Headers, []byte("Content-Length: "+bodyLen))
	}

	return response
}
