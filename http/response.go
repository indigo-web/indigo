package http

import (
	"indigo/internal"
)

func RenderHTTPResponse(
	buff []byte,
	proto []byte,
	code []byte,
	status []byte,
	headers Headers,
	body []byte) []byte {

	buff = append(append(append(buff, proto...), code...), status...)

	for key, value := range headers {
		buff = append(append(buff, internal.S2B(key)...), value...)
	}

	return append(buff, body...)
}
