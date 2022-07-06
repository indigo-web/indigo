package types

import (
	"indigo/http"
	"indigo/internal"
)

type ResponseStruct struct {
	Code    http.StatusCode
	Headers http.Headers
	Body    []byte
}

func Response() *ResponseStruct {
	return &ResponseStruct{}
}

func (r *ResponseStruct) WithCode(code http.StatusCode) *ResponseStruct {
	r.Code = code
	return r
}

func (r *ResponseStruct) WithHeader(key, value string) *ResponseStruct {
	r.Headers[key] = internal.S2B(value)
	return r
}

func (r *ResponseStruct) WithBody(body []byte) *ResponseStruct {
	r.Body = body
	return r
}

func (r *ResponseStruct) WithBodyString(body string) *ResponseStruct {
	r.Body = internal.S2B(body)
	return r
}
