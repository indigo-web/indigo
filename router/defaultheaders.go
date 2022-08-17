package router

import (
	"indigo/types"
	"strings"
)

func (d DefaultRouter) SetDefaultHeaders(headers types.ResponseHeaders) {
	d.renderer.SetDefaultHeaders(headers)
}

func (d DefaultRouter) applyDefaultHeaders() {
	encodings := strings.Join(d.codings.Acceptable(), ", ")

	d.renderer.SetDefaultHeaders(types.ResponseHeaders{
		"Server":          "indigo",
		"Connection":      "keep-alive",
		"Accept-Encoding": encodings,
	})
}
