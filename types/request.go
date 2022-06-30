package types

import (
	"indigo/http"
	"indigo/internal"
)

type (
	Params map[string][]byte
)

type Request struct {
	Method   http.Method
	Path     []byte
	Params   Params
	Protocol http.ProtocolVersion
	Headers  http.Headers

	body requestBody
}

func NewRequest(pathBuffer []byte, headers http.Headers, params Params) (Request, func([]byte)) {
	pipe := *internal.NewPipe()

	return Request{
		Path:     pathBuffer,
		Params:   params,
		Protocol: 0,
		Headers:  headers,
		body: requestBody{
			body: pipe,
		},
	}, pipe.Write
}

func (r *Request) GetBody(bodyCb onBodyCallback, completeCb onBodyCompleteCallback) error {
	return r.body.Read(bodyCb, completeCb)
}
