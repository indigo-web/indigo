package inbuilt

import "indigo/http/encodings"

/*
This file is responsible for adding content-encodings decoders/encoders
by user, so like custom content-encodings is a way of usage

At the moment is unused because no ideas how to implement it (why tf router
must do this lol, he does not even access anything for this)
*/

// AddContentEncoding simply adds a new custom decoder
func (d DefaultRouter) AddContentEncoding(token string, decoder encodings.Decoder) {
	d.codings.AddDecoder(token, decoder)
}
