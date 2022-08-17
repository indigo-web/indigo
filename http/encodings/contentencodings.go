package encodings

import (
	"indigo/http/render"
	"indigo/internal"
	"strings"
)

type (
	Decoder func([]byte) []byte
)

type ContentEncodings struct {
	encodings             map[string]Decoder
	acceptEncodingsHeader []byte
}

func NewContentEncodings() ContentEncodings {
	return ContentEncodings{
		encodings:             make(map[string]Decoder),
		acceptEncodingsHeader: []byte("Accept-Encodings: identity"),
	}
}

func (c ContentEncodings) GetDecoder(token string) (decoder Decoder, found bool) {
	decoder, found = c.encodings[token]
	return decoder, found
}

func (c ContentEncodings) AddDecoder(token string, decoder Decoder) {
	c.encodings[token] = decoder
	acceptable := strings.Join(internal.Keys(c.encodings), ", ")
	c.acceptEncodingsHeader = render.Header("Accept-Encodings", acceptable)
}

func (c ContentEncodings) Acceptable() []byte {
	return c.acceptEncodingsHeader
}
