package http

import "indigo/internal"

type Method uint8

const (
	GET Method = iota + 1
	HEAD
	POST
	PUT
	DELETE
	CONNECT
	OPTIONS
	TRACE
	PATCH
)

func GetMethod(method []byte) Method {
	switch internal.B2S(method) {
	case "GET":
		return GET
	case "HEAD":
		return HEAD
	case "POST":
		return POST
	case "PUT":
		return PUT
	case "DELETE":
		return DELETE
	case "CONNECT":
		return CONNECT
	case "OPTIONS":
		return OPTIONS
	case "TRACE":
		return TRACE
	case "PATCH":
		return PATCH
	}

	return 0
}

func Method2String(method Method) string {
	switch method {
	case GET:
		return "GET"
	case HEAD:
		return "HEAD"
	case POST:
		return "POST"
	case PUT:
		return "PUT"
	case DELETE:
		return "DELETE"
	case CONNECT:
		return "CONNECT"
	case OPTIONS:
		return "OPTIONS"
	case TRACE:
		return "TRACE"
	case PATCH:
		return "PATCH"
	}

	return "UNKNOWN"
}
