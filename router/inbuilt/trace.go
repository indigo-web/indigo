package inbuilt

import (
	"strings"

	"github.com/indigo-web/indigo/http"
)

/*
This file is responsible for rendering http requests. Prime use case is rendering
http requests back as a response to a trace request
*/

func traceResponse(respond *http.Response, messageBody []byte) *http.Response {
	return respond.
		Header("Content-Type", "message/http").
		Bytes(messageBody)
}

func renderHTTPRequest(request *http.Request, buff []byte) []byte {
	buff = append(buff, request.Method.String()...)
	buff = append(buff, ' ')
	buff = requestURI(request, buff)
	buff = append(buff, ' ')
	buff = append(buff, strings.TrimSpace(request.Protocol.String())...)
	buff = append(buff, "\r\n"...)
	buff = requestHeaders(request.Headers, buff)
	buff = append(buff, "Content-Length: 0\r\n\r\n"...)

	return buff
}

func requestURI(request *http.Request, buff []byte) []byte {
	buff = append(buff, request.Path...)
	buff = requestURIParams(request.Params, buff)

	return buff
}

func requestURIParams(params http.Params, buff []byte) []byte {
	if params.Empty() {
		return buff
	}

	buff = append(buff, '?')

	for key, val := range params.Pairs() {
		buff = append(buff, key...)
		if len(val) > 0 {
			buff = append(buff, '=')
			buff = append(buff, val...)
		}

		buff = append(buff, '&')
	}

	return buff[:len(buff)-1]
}

func requestHeaders(hdrs http.Headers, buff []byte) []byte {
	for _, pair := range hdrs.Expose() {
		buff = append(append(buff, pair.Key...), ": "...)
		buff = append(append(buff, pair.Value...), "\r\n"...)
	}

	return buff
}
