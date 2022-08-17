package router

import "indigo/http/encodings"

func (d DefaultRouter) AddContentEncoding(token string, decoder encodings.Decoder) {
	d.codings.AddDecoder(token, decoder)
}
