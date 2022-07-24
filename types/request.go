package types

import (
	"indigo/http"
	"indigo/internal"
)

type (
	Params map[string][]byte
)

// DefaultBodySize TODO: set this value in settings
const DefaultBodySize = 1024

type Request struct {
	Method   http.Method
	Path     []byte
	Params   Params
	Protocol http.Protocol
	Headers  http.Headers

	body requestBody
}

func NewRequest(pathBuffer []byte, headers http.Headers, params Params) (Request, *internal.Pipe) {
	// pipe is sized chan because parser can write an error even before
	// handler will be called
	pipe := internal.NewChanSizedPipe(0, 1)

	return Request{
		Path:     pathBuffer,
		Params:   params,
		Protocol: http.Protocol{},
		Headers:  headers,
		body: requestBody{
			body: pipe,
		},
	}, pipe
}

func (r *Request) GetBody(bodyCb onBodyCallback, completeCb onBodyCompleteCallback) error {
	return r.body.Read(bodyCb, completeCb)
}

func (r *Request) GetFullBody() []byte {
	buffer := make([]byte, 0, DefaultBodySize)

	// currently we have no limit for max body size
	// so no error is possible to occur
	_ = r.GetBody(
		func(b []byte) error {
			buffer = append(buffer, b...)
			return nil
		},
		func(err error) {
			return
		},
	)

	return buffer
}

func (r *Request) Reset() {
	// TODO: headers must also be reset
	r.body.Reset()
}
