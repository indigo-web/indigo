package codec

import (
	"github.com/klauspost/compress/gzip"
)

func NewGZIP() Codec {
	writer := gzip.NewWriter(nil)
	reader := new(gzip.Reader)
	instantiator := newBaseInstance(writer, reader, genericResetter)

	return newBaseCodec("gzip", instantiator)
}
