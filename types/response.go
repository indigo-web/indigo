package types

import (
	"indigo/http"
	"indigo/internal"
)

var (
	headerKeyValueSplitter = []byte(": ")
	crlf                   = []byte("\r\n")
)

type Response struct {
	Code    http.StatusCode
	Headers [][]byte
	Body    []byte
}

func NewResponse() Response {
	return Response{}
}

func (r Response) WithCode(code http.StatusCode) Response {
	r.Code = code
	return r
}

func (r Response) WithHeaderByte(key, value []byte) Response {
	r.Headers = append(r.Headers, renderHeader(key, value))
	return r
}

func (r Response) WithHeader(key, value string) Response {
	return r.WithHeaderByte(internal.S2B(key), internal.S2B(value))
}

func (r Response) WithBodyByte(body []byte) Response {
	r.Body = append(r.Body, body...)
	return r
}

func (r Response) WithBody(body string) Response {
	return r.WithBodyByte(internal.S2B(body))
}

func (r Response) GrowHeaders(newsize int) Response {
	if cap(r.Headers) < newsize {
		r.Headers = append(make([][]byte, 0, newsize), r.Headers...)
	}

	return r
}

/*
render function takes buffer (that already starts with protocol and space),
rendering and writing response directly into the response writer
*/
func (r Response) render(buff []byte, writer ResponseWriter) (err error, modifiedBuff []byte) {
	// append code and status to buff
	buff = append(append(buff, http.GetByteCodeTrailingSpace(r.Code)...), http.GetStatusTrailingCRLF(r.Code)...)

	for _, header := range r.Headers {
		buff = append(append(buff, header...), crlf...)
	}

	buff = append(buff, crlf...)

	return writer(append(buff, r.Body...)), buff
}

func renderHeader(key, value []byte) []byte {
	return append(append(key, headerKeyValueSplitter...), value...)
}

func WriteResponse(buff []byte, response Response, writer ResponseWriter) (error, []byte) {
	return response.render(buff, writer)
}
