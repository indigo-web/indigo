package serialize

import (
	"strconv"

	"github.com/indigo-web/indigo/http"
)

func Headers(request *http.Request) string {
	var buff []byte

	buff = append(buff, request.Method.String()...)
	buff = append(buff, ' ')
	buff = append(buff, request.Path...)

	if !request.Params.Empty() {
		buff = append(buff, '?')

		for key, value := range request.Params.Pairs() {
			buff = queryparam(buff, key, value)
			buff = append(buff, '&')
		}

		if len(buff) > 0 && buff[len(buff)-1] == '&' {
			buff = buff[:len(buff)-1]
		}
	}

	buff = append(buff, ' ')
	buff = append(buff, request.Protocol.String()...)
	buff = append(buff, '\r', '\n')

	for _, h := range request.Headers.Expose() {
		buff = header(buff, h)
	}

	if !request.Encoding.Chunked {
		buff = header(buff, http.Header{
			Key:   "Content-Length",
			Value: strconv.Itoa(request.ContentLength),
		})
	}

	buff = append(buff, '\r', '\n')

	return string(buff)
}

func Request(request *http.Request) (string, error) {
	headers := Headers(request)
	body, err := request.Body.String()

	return headers + body, err
}

func header(b []byte, h http.Header) []byte {
	b = append(b, h.Key...)
	b = append(b, ':', ' ')
	b = append(b, h.Value...)

	return append(b, '\r', '\n')
}

func queryparam(buff []byte, key, value string) []byte {
	buff = append(buff, key...)
	buff = append(buff, '=')
	buff = append(buff, value...)
	return buff
}
