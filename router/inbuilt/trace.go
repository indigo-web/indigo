package inbuilt

import (
	"github.com/indigo-web/indigo/http"
	"github.com/indigo-web/indigo/http/headers"
	"github.com/indigo-web/indigo/internal/httpchars"
	"strings"
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
	buff = append(buff, httpchars.SP...)
	buff = requestURI(request, buff)
	buff = append(buff, httpchars.SP...)
	buff = append(buff, strings.TrimSpace(request.Proto.String())...)
	buff = append(buff, httpchars.CRLF...)
	buff = requestHeaders(request.Headers, buff)
	buff = append(buff, "Content-Length: 0\r\n\r\n"...)

	return buff
}

func requestURI(request *http.Request, buff []byte) []byte {
	buff = append(buff, request.Path...)

	if query := request.Query.Raw(); len(query) > 0 {
		buff = append(buff, '?')
		buff = append(buff, query...)
	}

	return buff
}

func requestHeaders(hdrs headers.Headers, buff []byte) []byte {
	for _, pair := range hdrs.Unwrap() {
		buff = append(append(buff, pair.Key...), httpchars.COLONSP...)
		buff = append(append(buff, pair.Value...), httpchars.CRLF...)
	}

	return buff
}
