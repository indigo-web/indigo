package codec

import (
	"io"

	"github.com/klauspost/compress/flate"
)

func NewDeflate() Codec {
	writer, err := flate.NewWriter(nil, 5)
	if err != nil {
		panic(err)
	}

	reader := flate.NewReader(nil)
	instantiator := newBaseInstance(writer, reader, func(r io.Reader, a *readerAdapter) error {
		return r.(flate.Resetter).Reset(a, nil)
	})

	return newBaseCodec("deflate", instantiator)
}
