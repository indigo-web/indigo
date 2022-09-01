package inbuilt

import (
	"github.com/fakefloordiv/indigo/http/headers"
	"strings"
)

var (
	server     = []byte("indigo")
	connection = []byte("keep-alive")
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
			"Server":          server,
			"Connection":      connection,
			"Accept-Encoding": []byte(encodings),
		})
	}

	d.renderer.SetDefaultHeaders(d.defaultHeaders)
}
