package codec

import (
	"io"

	"github.com/indigo-web/indigo/http"
)

type Codec interface {
	// Token returns a coding token associated with the codec itself.
	Token() string
	New() Instance
}

type Instance interface {
	Compressor
	Decompressor
}

type Compressor interface {
	ResetCompressor(w io.Writer)
	io.Writer
	Flush() error
}

type Decompressor interface {
	ResetDecompressor(source http.Fetcher) error
	http.Fetcher
}
