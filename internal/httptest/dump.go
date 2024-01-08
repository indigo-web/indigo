package httptest

import (
	"fmt"
	"github.com/indigo-web/indigo/http"
	"github.com/indigo-web/indigo/http/headers"
	"github.com/indigo-web/indigo/http/proto"
)

func Dump(request *http.Request) (string, error) {
	var buff []byte

	buff = append(buff, request.Method.String()...)
	buff = space(buff)
	buff = append(buff, request.Path...)

	if raw := request.Query.Raw(); len(raw) > 0 {
		buff = question(buff)
		buff = append(buff, raw...)
	}

	buff = space(buff)
	protocol := proto.ToBytes(request.Proto)
	protocol = protocol[:len(protocol)-1]
	buff = append(buff, protocol...)
	buff = crlf(buff)

	for _, h := range request.Headers.Unwrap() {
		buff = header(buff, h)
	}

	buff = crlf(buff)
	fmt.Println("???")
	body, err := request.Body.Bytes()
	buff = append(buff, body...)

	return string(buff), err
}

func space(b []byte) []byte {
	return append(b, ' ')
}

func question(b []byte) []byte {
	return append(b, '?')
}

func crlf(b []byte) []byte {
	return append(b, '\r', '\n')
}

func header(b []byte, h headers.Header) []byte {
	b = append(b, h.Key...)
	b = colonsp(b)
	b = append(b, h.Value...)

	return crlf(b)
}

func colonsp(b []byte) []byte {
	return append(b, ':', ' ')
}
