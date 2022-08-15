package render

import (
	"indigo/http/proto"
	"indigo/http/status"
	"indigo/types"
	"strconv"
)

var (
	space           = []byte(" ")
	crlf            = []byte("\r\n")
	colonSpace      = []byte(": ")
	headerSplitters = [2][]byte{colonSpace, crlf}
	contentLength   = []byte("Content-Length: ")
)

type Renderer struct {
	buff []byte
}

func NewRenderer(buff []byte) Renderer {
	return Renderer{
		buff: buff,
	}
}

func (r *Renderer) Response(protocol proto.Proto, response types.Response) []byte {
	buff := r.buff[:0]
	buff = append(append(buff, proto.ToBytes(protocol)...), space...)
	buff = append(append(buff, strconv.Itoa(int(response.Code))...), space...)
	buff = append(append(buff, status.Text(response.Code)...), crlf...)

	headers := response.Headers()

	for i := range headers {
		buff = append(append(buff, headers[i]...), headerSplitters[i%2]...)
	}

	buff = append(append(append(buff, contentLength...), strconv.Itoa(len(response.Body))...), crlf...)
	r.buff = append(append(buff, crlf...), response.Body...)

	return r.buff
}
