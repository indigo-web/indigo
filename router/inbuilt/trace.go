package inbuilt

import (
	"github.com/fakefloordiv/indigo/http"
	"github.com/fakefloordiv/indigo/http/headers"
	methods "github.com/fakefloordiv/indigo/http/method"
	"github.com/fakefloordiv/indigo/http/parser/http1"
	"github.com/fakefloordiv/indigo/http/proto"
	"github.com/fakefloordiv/indigo/internal"
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
	for k, v := range request.Headers {
		buff = append(buff, k...)
		buff = append(append(buff, http.COLON...), http.SP...)
		buff = renderHeaderValues(v, buff)
		buff = append(buff, http.CRLF...)
	}

	return buff
}

func renderHeaderValues(values []headers.Header, buff []byte) []byte {
	for i := range values[:len(values)-1] {
		buff = renderHeaderValue(values[i], buff)
		buff = append(buff, http.COMMA...)
	}

	return renderHeaderValue(values[len(values)-1], buff)
}

func renderHeaderValue(value headers.Header, buff []byte) []byte {
	buff = append(buff, value.Value...)

	if value.Q != 10 {
		buff = append(append(buff, ";q=0."...), value.QualityString()...)
	}

	if value.Charset != internal.B2S(http1.DefaultCharset) {
		buff = append(append(buff, ";charset="...), value.Charset...)
	}

	return buff
}
