package inbuilt

import (
	"bytes"
	"github.com/indigo-web/indigo/http"
	methods "github.com/indigo-web/indigo/http/method"
	"github.com/indigo-web/indigo/http/proto"
	"github.com/indigo-web/indigo/internal/httpchars"
)

/*
This file is responsible for rendering http requests. Prime use case is rendering
http requests back as a response to a trace request
*/

func traceResponse(respond http.Response, messageBody []byte) http.Response {
	return respond.
		WithHeader("Content-Type", "message/http").
		WithBodyByte(messageBody)
}

func renderHTTPRequest(request *http.Request, buff []byte) []byte {
	buff = append(buff, methods.ToString(request.Method)...)
	buff = append(buff, httpchars.SP...)
	buff = requestURI(request, buff)
	buff = append(buff, httpchars.SP...)
	buff = append(buff, bytes.TrimSpace(proto.ToBytes(request.Proto))...)
	buff = append(buff, httpchars.CRLF...)
	buff = requestHeaders(request, buff)
	buff = append(buff, "content-length: 0\r\n"...)

	return append(buff, httpchars.CRLF...)
}

func requestURI(request *http.Request, buff []byte) []byte {
	buff = append(buff, request.Path...)

	if query := request.Query.Raw(); len(query) > 0 {
		buff = append(buff, '?')
		buff = append(buff, query...)
	}

	if len(request.Fragment) > 0 {
		buff = append(buff, '#')
		buff = append(buff, request.Fragment...)
	}

	return buff
}

func requestHeaders(request *http.Request, buff []byte) []byte {
	for k, v := range request.Headers.Unwrap() {
		buff = append(append(buff, k...), httpchars.COLONSP...)
		buff = joinValuesInto(buff, v)
	}

	return buff
}

func joinValuesInto(buff []byte, values []string) []byte {
	for i := range values[:len(values)-1] {
		buff = append(append(buff, values[i]...), httpchars.COMMA...)
	}

	return append(append(buff, values[len(values)-1]...), httpchars.CRLF...)
}
