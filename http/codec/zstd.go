package codec

import (
	"github.com/klauspost/compress/zstd"
)

func NewZSTD() Codec {
	w, err := zstd.NewWriter(nil)
	if err != nil {
		panic(err)
	}

	r, err := zstd.NewReader(nil)
	if err != nil {
		panic(err)
	}

	instantiator := newBaseInstance(w, r, genericResetter)

	return newBaseCodec("zstd", instantiator)
}
