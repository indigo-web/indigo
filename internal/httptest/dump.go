package httptest

import (
	"github.com/indigo-web/indigo/http"
	"github.com/indigo-web/indigo/http/headers"
	"strconv"
)

func Dump(request *http.Request) (string, error) {
	var buff []byte

	buff = append(buff, request.Method.String()...)
	buff = space(buff)
	buff = append(buff, request.Path...)

	if raw := request.Query.Raw(); len(raw) > 0 {
		buff = question(buff)
		params, err := request.Query.Unwrap()
		if err != nil {
			return "", err
		}

		for _, param := range params.Unwrap() {
			buff = queryparam(buff, param.Key, param.Value)
			buff = ampersand(buff)
		}

		if len(buff) > 0 && buff[len(buff)-1] == '&' {
			buff = buff[:len(buff)-1]
		}
	}

	buff = space(buff)
	protocol := request.Proto.String()
	protocol = protocol[:len(protocol)-1]
	buff = append(buff, protocol...)
	buff = crlf(buff)

	for _, h := range request.Headers.Unwrap() {
		buff = header(buff, h)
	}

	buff = header(buff, headers.Header{
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

func header(b []byte, h headers.Header) []byte {
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
