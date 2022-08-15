package types

import (
	"indigo/http/status"
	"indigo/internal"
)

type ResponseWriter func([]byte) error

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
	headers ResponseHeaders
	Body    []byte
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
	return r.WithHeaderByte(internal.S2B(key), internal.S2B(value))
}

func (r Response) WithHeaderByte(key, value []byte) Response {
	r.headers = r.headers.Append(key, value)
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

func (r Response) Headers() ResponseHeaders {
	return r.headers
}
