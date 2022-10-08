package inbuilt

import (
	"github.com/fakefloordiv/indigo/http"
	methods "github.com/fakefloordiv/indigo/http/method"
	"github.com/fakefloordiv/indigo/http/proto"
	"github.com/fakefloordiv/indigo/types"
)

/*
This file is responsible for rendering http requests. Prime use case is rendering
http requests back as a response to a trace request
*/

func traceResponse(messageBody []byte) types.Response {
	return types.
		WithHeader("Content-Type", "message/http").
		WithBodyByte(messageBody)
}

func renderHTTPRequest(request *types.Request, buff []byte) []byte {
	buff = append(buff, methods.ToString(request.Method)...)
	buff = append(buff, http.SP...)
	buff = requestURI(request, buff)
	buff = append(buff, http.SP...)
	buff = append(buff, proto.ToBytes(request.Proto)...)
	buff = append(buff, http.CRLF...)
	buff = requestHeaders(request, buff)
	buff = append(buff, "content-length: 0\r\n"...)

	return append(buff, http.CRLF...)
}

func requestURI(request *types.Request, buff []byte) []byte {
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

func requestHeaders(request *types.Request, buff []byte) []byte {
	for k, v := range request.Headers.AsMap() {
		buff = append(append(buff, k...), http.COLONSP...)
		buff = append(joinValuesInto(buff, v))
	}

	return buff
}

func joinValuesInto(buff []byte, values []string) []byte {
	for i := range values[:len(values)-1] {
		buff = append(append(buff, values[i]...), http.COMMA...)
	}

	return append(append(buff, values[len(values)-1]...), http.CRLF...)
}
