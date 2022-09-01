package encodings

import (
	"github.com/fakefloordiv/indigo/internal"
)

type (
	Decoder func([]byte) []byte
)

// ContentEncodings is just a structure that encapsulates containing
// content decoders. It does not much but it's honest work
type ContentEncodings struct {
	encodings map[string]Decoder
}

// NewContentEncodings returns new instance of ContentEncodings
func NewContentEncodings() ContentEncodings {
	return ContentEncodings{
		encodings: make(map[string]Decoder),
	}
}

// GetDecoder takes a string as an encoding token (name), returns
// corresponding Decoder
// in case x-gzip is passed, decoder for gzip is returned, see
// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Content-Encoding#directives
func (c ContentEncodings) GetDecoder(token string) (decoder Decoder, found bool) {
	if token == "x-gzip" {
		token = "gzip"
	}

	decoder, found = c.encodings[token]
	return decoder, found
}

// AddDecoder simply adds a new decoder
func (c ContentEncodings) AddDecoder(token string, decoder Decoder) {
	c.encodings[token] = decoder
}

// Acceptable returns a string with all the available decoders, listed
// by comma
func (c ContentEncodings) Acceptable() []string {
	if len(c.encodings) == 0 {
		return []string{"identity"}
	}

	return internal.Keys(c.encodings)
}
