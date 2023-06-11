package decode

import (
	"github.com/indigo-web/indigo/http/status"
	"github.com/indigo-web/indigo/internal/mapconv"
)

const identity = "identity"

type (
	// DecoderFactory returns a decoding function. It is being called only once per client,
	// and only if the client used it
	DecoderFactory interface {
		New() DecoderFunc
	}

	// DecoderFunc is a function that returns decoded data. Returned data may be empty
	// TODO: are there guarantees, that returned slice with uncompressed data won't be used
	//       even after the next call? Depending on this, size of the buffer may vary greatly
	DecoderFunc func(encoded []byte) (decoded []byte, err error)
)

// Decoder is the encapsulation of the mechanism of retrieving a decoder for each request.
// In case no encoding is used in the request (identity is specified, or not mentioned at all),
// original data will be returned
type Decoder struct {
	factories map[string]DecoderFactory
	keys      []string
	decoders  []DecoderFunc
}

func NewDecoder() *Decoder {
	return &Decoder{
		factories: make(map[string]DecoderFactory),
	}
}

// Decode decodes the input data, based on a token, where token is a key
func (d *Decoder) Decode(token string, encoded []byte) (decoded []byte, err error) {
	if len(token) == 0 || token == identity {
		return encoded, nil
	}

	// As we know, there usually aren't many decoders, so the cheapest way is just to brute
	// the slice with keys. This approach usually consumes at most 3-4ns in pretty loaded cases,
	// like 10-20 elements are provided. Pretty cheap
	for i := range d.keys {
		if token == d.keys[i] {
			return d.decoders[i](encoded)
		}
	}

	// no already spawned decoder was matched, so try to add a new one. This going to take a while
	// because of the mapaccess operation, but this appears not that often in the common case

	factory, ok := d.factories[token]
	if !ok {
		return nil, status.ErrNotImplemented
	}

	d.keys = append(d.keys, token)
	decoder := factory.New()
	d.decoders = append(d.decoders, decoder)

	return decoder(encoded)
}

func (d *Decoder) Acceptable() []string {
	if len(d.factories) == 0 {
		return []string{identity}
	}

	return mapconv.Keys(d.factories)
}

// Add adds a new decoder factory. With gzip and compress encoding tokens, also x-gzip
// and x-compress tokens will be automatically included
func (d *Decoder) Add(token string, factory DecoderFactory) {
	d.factories[token] = factory

	// this exists in backward-capability purposes. Some old clients may use x-gzip or
	// x-compress instead of regular gzip or compress tokens respectively.
	// see https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Content-Encoding#directives
	switch token {
	case "gzip":
		d.factories["x-gzip"] = factory
	case "compress":
		d.factories["x-compress"] = factory
	}
}
