package types

import (
	"indigo/http/headers"
	"indigo/http/status"
	"indigo/internal"
)

type (
	ResponseWriter func([]byte) error
)

// WithResponse is just a nil-filled default pre-created response. Because
// of clear methods, it is anyway copied every time it is used as constructor
// so please, DO NOT modify fields of this variable
var WithResponse = Response{
	Code:   status.OK,
	Status: status.Text(status.OK),
}

type Response struct {
	Code   status.Code
	Status status.Status
	// headers due to possible side effects are decided to be private
	// also uninitialized response must ALWAYS have this value as nil
	headers headers.Headers
	Body    []byte
}

func NewResponse() Response {
	return Response{
		Code:    status.OK,
		Status:  status.Text(status.OK),
		headers: make(headers.Headers),
	}
}

func (r Response) WithCode(code status.Code) Response {
	r.Code = code
	r.Status = status.Text(code)
	return r
}

func (r Response) WithStatus(status status.Status) Response {
	r.Status = status
	return r
}

func (r Response) WithHeader(key, value string) Response {
	if r.headers == nil {
		r.headers = headers.Headers{
			key: internal.S2B(value),
		}

		return r
	}

	r.headers[key] = internal.S2B(value)

	return r
}

func (r Response) WithBody(body string) Response {
	r.Body = internal.S2B(body)
	return r
}

func (r Response) WithBodyByte(body []byte) Response {
	r.Body = body
	return r
}

func (r Response) Headers() headers.Headers {
	return r.headers
}
