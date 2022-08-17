package render

import (
	"indigo/http/proto"
	"indigo/http/status"
	"indigo/types"
	"strconv"
)

var (
	space         = []byte(" ")
	crlf          = []byte("\r\n")
	colonSpace    = []byte(": ")
	contentLength = []byte("Content-Length: ")
)

type Renderer struct {
	buff []byte

	defaultHeaders types.ResponseHeaders
}

func NewRenderer(buff []byte) *Renderer {
	return &Renderer{
		buff: buff,
	}
}

func (r *Renderer) SetDefaultHeaders(headers types.ResponseHeaders) {
	r.defaultHeaders = headers
}

func (r *Renderer) Response(protocol proto.Proto, response types.Response) []byte {
	buff := r.buff[:0]
	buff = append(append(buff, proto.ToBytes(protocol)...), space...)
	buff = append(append(buff, strconv.Itoa(int(response.Code))...), space...)
	buff = append(append(buff, status.Text(response.Code)...), crlf...)

	headers := response.Headers()

	for key, value := range headers {
		buff = append(renderHeader(key, value, buff), crlf...)
	}

	for key, value := range r.defaultHeaders {
		_, found := headers[key]
		if !found {
			buff = append(renderHeader(key, value, buff), crlf...)
		}
	}

	buff = append(append(append(buff, contentLength...), strconv.Itoa(len(response.Body))...), crlf...)
	r.buff = append(append(buff, crlf...), response.Body...)

	return r.buff
}

func renderHeader(key, value string, into []byte) []byte {
	return append(append(append(into, key...), colonSpace...), value...)
}
