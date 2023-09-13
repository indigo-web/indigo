package client

import (
	"github.com/indigo-web/indigo/http"
	"github.com/indigo-web/indigo/http/headers"
	"github.com/indigo-web/indigo/http/proto"
	"github.com/indigo-web/indigo/http/status"
)

type Response struct {
	Protocol      proto.Proto
	Code          status.Code
	Status        status.Status
	Headers       *headers.Headers
	ContentLength int
	Encoding      headers.Encoding
	Body          *http.Body
}

func NewResponse(headers *headers.Headers, body *http.Body) Response {
	return Response{
		Headers: headers,
		Body:    body,
	}
}
