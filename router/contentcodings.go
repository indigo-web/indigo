package router

import "indigo/http/encodings"

/*
This file is responsible for adding content-encodings decoders/encoders
by user, so like custom content-encodings is a way of usage
*/

// AddContentEncoding simply adds a new custom decoder
func (d DefaultRouter) AddContentEncoding(token string, decoder encodings.Decoder) {
	d.codings.AddDecoder(token, decoder)
}
