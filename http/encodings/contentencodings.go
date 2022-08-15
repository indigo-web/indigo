package encodings

type (
	Decoder         func([]byte) []byte
	token           string
	AcceptEncodings []byte
)

type ContentEncodings struct {
	encodings             map[token]Decoder
	acceptEncodingsHeader AcceptEncodings
}

func NewContentEncodings() ContentEncodings {
	return ContentEncodings{
		encodings:             make(map[token]Decoder),
		acceptEncodingsHeader: []byte("Accept-Encodings: identity"),
	}
}

func (c ContentEncodings) GetDecoder(tok token) (decoder Decoder, found bool) {
	decoder, found = c.encodings[tok]
	return decoder, found
}

func (c ContentEncodings) AddDecoder(tok token, decoder Decoder) {
	c.encodings[tok] = decoder
}

func (c ContentEncodings) Acceptable() AcceptEncodings {
	return c.acceptEncodingsHeader
}
