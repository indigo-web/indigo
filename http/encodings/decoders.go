package encodings

import (
	"github.com/fakefloordiv/indigo/internal/mapconv"
)

type (
	// Decoder is simply a decoder factory that returns a new decoder func
	// for every request (but only in case it is encoded)
	Decoder interface {
		New() DecoderFunc
	}

	// DecoderFunc is a function that takes encoded byte-slice and returns
	// decoded one
	DecoderFunc func(encoded []byte) (decoded []byte, err error)
)

// Decoders structure just encapsulates containing content decoders.
// This used for some http-compatibility actions like backward capability
// for compress and x-compress, gzip and x-gzip
type Decoders struct {
	decoders map[string]Decoder
}

// NewContentDecoders returns new instance of Decoders
func NewContentDecoders() Decoders {
	return Decoders{
		decoders: make(map[string]Decoder),
	}
}

// Get takes a string as an encoding token (name), returns
// corresponding Decoder
func (d Decoders) Get(token string) (decoder Decoder, found bool) {
	decoder, found = d.decoders[token]
	return decoder, found
}

// GetDecoder is an ordinary Get, still exists in backward-capability purposes
func (d Decoders) GetDecoder(token string) (decoder Decoder, found bool) {
	return d.Get(token)
}

// Add simply adds a new decoder. In case gzip or compress is
// passed also x-gzip and x-compress keys will be automatically appended,
// see https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Content-Encoding#directives
func (d Decoders) Add(token string, decoder Decoder) {
	d.decoders[token] = decoder

	switch token {
	case "gzip":
		d.decoders["x-gzip"] = decoder
	case "compress":
		d.decoders["x-compress"] = decoder
	}
}

// AddDecoder does the same as Add except it adds directly decoderFunc. Left mostly in
// backward capability purposes
func (d Decoders) AddDecoder(token string, decoderFunc DecoderFunc) {
	d.Add(token, newNopDecoder(decoderFunc))
}

// Acceptable returns a string with all the available decoders, listed
// by comma. In case no decoders are presented, identity is used to notify
// a client that server does not accept any encodings
func (d Decoders) Acceptable() []string {
	if len(d.decoders) == 0 {
		return []string{"identity"}
	}

	return mapconv.Keys(d.decoders)
}

type nopDecoder struct {
	decoderFunc DecoderFunc
}

func newNopDecoder(decoderFunc DecoderFunc) Decoder {
	return nopDecoder{
		decoderFunc: decoderFunc,
	}
}

func (n nopDecoder) New() DecoderFunc {
	return n.decoderFunc
}
