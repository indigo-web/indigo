package inbuilt

import (
	"strings"

	"github.com/fakefloordiv/indigo/http/headers"
)

const (
	server      = "indigo"
	connection  = "keep-alive"
	contentType = "text/html"
)

// SetDefaultHeaders sets headers by default. This action overrides already set
// default headers, and cause of that, Accept-Encoding header may be not
// behave correctly. So using this method is unwanted, but if you do, use with
// care
func (d *DefaultRouter) SetDefaultHeaders(headers headers.Headers) {
	d.defaultHeaders = headers
}

// applyDefaultHeaders is called when server is initialized and ready to work
// the only thing it does is setting headers by default if they were not set
// before
func (d *DefaultRouter) applyDefaultHeaders() {
	if d.defaultHeaders == nil {
		encodings := strings.Join(d.codings.Acceptable(), ", ")

		d.SetDefaultHeaders(headers.Headers{
			// not specifying version to avoid problems with vulnerable versions
			// Also not specifying, because otherwise I should add an option to
			// disable such a behaviour. I am too lazy for that. Preferring not
			// specifying version at all
			"Server": server,
			// Automatically specify connection as keep-alive. Maybe it is not
			// a compulsory move from server, but still
			"Connection": connection,
			// explicitly set a list of encodings are supported to avoid awkward
			// situations
			"Accept-Encoding": encodings,
			// content-type that is set by-default. In case you set a custom one
			// this will be overridden
			"Content-Type": contentType,
		})
	}

	d.renderer.SetDefaultHeaders(d.defaultHeaders)
}
