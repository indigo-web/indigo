package testutil

import (
	"github.com/indigo-web/indigo/http"
	"strconv"
)

func SerializeRequest(request *http.Request) (string, error) {
	var buff []byte

	buff = append(buff, request.Method.String()...)
	buff = space(buff)
	buff = append(buff, request.Path...)

	if !request.Params.Empty() {
		buff = question(buff)

		for key, value := range request.Params.Iter() {
			buff = queryparam(buff, key, value)
			buff = ampersand(buff)
		}

		if len(buff) > 0 && buff[len(buff)-1] == '&' {
			buff = buff[:len(buff)-1]
		}
	}

	buff = space(buff)
	buff = append(buff, request.Protocol.String()...)
	buff = crlf(buff)

	for _, h := range request.Headers.Expose() {
		buff = header(buff, h)
	}

	buff = header(buff, http.Header{
		Key:   "Content-Length",
		Value: strconv.Itoa(request.ContentLength),
	})

	buff = crlf(buff)
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

func ampersand(b []byte) []byte {
	return append(b, '&')
}

func crlf(b []byte) []byte {
	return append(b, '\r', '\n')
}

func header(b []byte, h http.Header) []byte {
	b = append(b, h.Key...)
	b = colonsp(b)
	b = append(b, h.Value...)

	return crlf(b)
}

func queryparam(buff []byte, key, value string) []byte {
	buff = append(buff, key...)
	buff = append(buff, '=')
	buff = append(buff, value...)
	return buff
}

func colonsp(b []byte) []byte {
	return append(b, ':', ' ')
}
