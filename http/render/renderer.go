package render

import (
	"github.com/fakefloordiv/indigo/http/headers"
	"github.com/fakefloordiv/indigo/http/proto"
	"github.com/fakefloordiv/indigo/http/status"
	"github.com/fakefloordiv/indigo/types"
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

// Response method is rendering types.Response object into some buffer and then writes
// it into the writer. Response method must provide next functionality:
// 1) Render types.Response object according to the provided protocol version
// 2) Be sure that used features of response are supported by client (provided protocol)
// 3) Support default headers
// 4) Add system-important headers, e.g. Content-Length
// 5) Content encodings must be applied here
// 6) Stream-based files uploading must be supported
func (r *Renderer) Response(
	protocol proto.Proto, response types.Response, writer types.ResponseWriter,
) error {
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

	return writer(r.buff)
}

func renderHeader(key string, value []byte, into []byte) []byte {
	return append(append(append(into, key...), colonSpace...), value...)
}
