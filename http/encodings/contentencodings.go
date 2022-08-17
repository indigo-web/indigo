package encodings

import (
	"indigo/internal"
)

type (
	Decoder func([]byte) []byte
)

type ContentEncodings struct {
	encodings map[string]Decoder
}

func NewContentEncodings() ContentEncodings {
	return ContentEncodings{
		encodings: make(map[string]Decoder),
	}
}

func (c ContentEncodings) GetDecoder(token string) (decoder Decoder, found bool) {
	decoder, found = c.encodings[token]
	return decoder, found
}

func (c ContentEncodings) AddDecoder(token string, decoder Decoder) {
	c.encodings[token] = decoder
}

func (c ContentEncodings) Acceptable() []string {
	if len(c.encodings) == 0 {
		return []string{"identity"}
	}

	return internal.Keys(c.encodings)
}
