package inbuilt

import (
	"indigo/http/status"
	"indigo/types"
)

/*
This file is responsible for default responses for errors. Each has its
own status code and text
*/

var (
	defaultBadRequest = types.WithResponse.
				WithCode(status.BadRequest).
				WithBody(`<h1 align="center">400 Bad Request</h1>`)

	defaultNotFound = types.WithResponse.
			WithCode(status.NotFound).
			WithBody(`<h1 align="center">404 Request Page Not Found</h1>`)

	defaultMethodNotAllowed = types.WithResponse.
				WithCode(status.MethodNotAllowed).
				WithBody(`<h1 align="center">405 Method Not Allowed</h1>`)

	defaultRequestEntityTooLarge = types.WithResponse.
					WithCode(status.RequestEntityTooLarge).
					WithBody(`<h1 align="center">413 Request Entity Too Large</h1>`)

	defaultURITooLong = types.WithResponse.
				WithCode(status.RequestURITooLong).
				WithBody(`<h1 align="center">414 Request URI Too Long</h1>`)

	defaultHeaderFieldsTooLarge = types.WithResponse.
					WithCode(status.RequestHeaderFieldsTooLarge).
					WithBody(`<h1 align="center">431 Request Header Fields Too Large</h1>`)

	defaultUnsupportedProtocol = types.WithResponse.
					WithCode(status.HTTPVersionNotSupported).
					WithBody(`<h1 align="center">505 HTTP Version Not Supported</h1>`)
)

type (
	ErrorHandler func(request *types.Request) types.Response
	errHandlers  map[status.Code]ErrorHandler
)

func newErrHandlers() errHandlers {
	return errHandlers{
		status.BadRequest:                  defaultBadRequestHandler,
		status.NotFound:                    defaultNotFoundHandler,
		status.MethodNotAllowed:            defaultMethodNotAllowedHandler,
		status.RequestEntityTooLarge:       defaultRequestEntityTooLargeHandler,
		status.ConnectionClose:             defaultConnectionClose,
		status.RequestURITooLong:           defaultURITooLongHandler,
		status.RequestHeaderFieldsTooLarge: defaultHeaderFieldsTooLargeHandler,
		status.HTTPVersionNotSupported:     defaultUnsupportedProtocolHandler,
	}
}

func defaultBadRequestHandler(_ *types.Request) types.Response {
	return defaultBadRequest
}

func defaultNotFoundHandler(_ *types.Request) types.Response {
	return defaultNotFound
}

func defaultMethodNotAllowedHandler(_ *types.Request) types.Response {
	return defaultMethodNotAllowed
}

func defaultRequestEntityTooLargeHandler(_ *types.Request) types.Response {
	return defaultRequestEntityTooLarge
}

func defaultConnectionClose(_ *types.Request) types.Response {
	// this is a special type of error handlers. Any response you return
	// will not be sent or anything will be done with it because calling
	// this handler means client has already disconnected
	return types.WithResponse
}

func defaultURITooLongHandler(_ *types.Request) types.Response {
	return defaultURITooLong
}

func defaultHeaderFieldsTooLargeHandler(_ *types.Request) types.Response {
	return defaultHeaderFieldsTooLarge
}

func defaultUnsupportedProtocolHandler(_ *types.Request) types.Response {
	return defaultUnsupportedProtocol
}
