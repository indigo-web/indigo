package render

import (
	"indigo/http/headers"
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

// Renderer is a session responses renderer. Its purpose is only to know
// something about client, and knowing them, render correct response
// for example, we SHOULD not use content-codings for HTTP/1.0 clients,
// and MUST NOT use them for HTTP/0.9 clients
// Also in case of file is being sent, it collects some meta about it,
// and compresses using available on both server and client encoders
type Renderer struct {
	buff []byte

	defaultHeaders headers.Headers
}

func NewRenderer(buff []byte) *Renderer {
	return &Renderer{
		buff: buff,
	}
}

func (r *Renderer) SetDefaultHeaders(headers headers.Headers) {
	r.defaultHeaders = headers
}

// Response provides next functionality:
// 1) rendering request into reusable buffer kept inside
// 2) rendering request includes: protocol, response code, response status,
//    headers, and body
// 3) default headers are also included. Default headers are just headers
//    that are being set only if they are not presented in user-headers
// 4) hard-coded content-length header. It must not be presented in user-headers,
//    otherwise client may disconnect with error, or even worse, this will cause
//    UB on client
func (r *Renderer) Response(protocol proto.Proto, response types.Response) []byte {
	buff := r.buff[:0]
	buff = append(append(buff, proto.ToBytes(protocol)...), space...)
	buff = append(append(buff, strconv.Itoa(int(response.Code))...), space...)
	buff = append(append(buff, status.Text(response.Code)...), crlf...)

	reqHeaders := response.Headers()

	for key, value := range reqHeaders {
		buff = append(renderHeader(key, value, buff), crlf...)
	}

	for key, value := range r.defaultHeaders {
		_, found := reqHeaders[key]
		if !found {
			buff = append(renderHeader(key, value, buff), crlf...)
		}
	}

	buff = append(append(append(buff, contentLength...), strconv.Itoa(len(response.Body))...), crlf...)
	r.buff = append(append(buff, crlf...), response.Body...)

	return r.buff
}

func renderHeader(key string, value []byte, into []byte) []byte {
	return append(append(append(into, key...), colonSpace...), value...)
}
