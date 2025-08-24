package codec

import (
	"github.com/klauspost/compress/gzip"
)

// TODO: pass this via parameters?
const gzipBufferSize = 4096

func NewGZIP() Codec {
	writer := gzip.NewWriter(nil)
	reader := new(gzip.Reader)
	instantiator := newBaseInstance(writer, reader, genericResetter)

	return newBaseCodec("gzip", gzipBufferSize, instantiator)
}
