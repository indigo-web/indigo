package types

import (
	"indigo/http"
	"indigo/internal"
)

type Response struct {
	Code    http.StatusCode
	Headers http.Headers
	Body    []byte
}

func (r *Response) WithCode(code http.StatusCode) *Response {
	r.Code = code
	return r
}

func (r *Response) WithHeader(key, value string) *Response {
	r.Headers[key] = internal.S2B(value)
	return r
}

func (r *Response) WithBody(body []byte) *Response {
	r.Body = body
	return r
}

func (r *Response) WithBodyString(body string) *Response {
	r.Body = internal.S2B(body)
	return r
}
