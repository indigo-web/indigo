package codec

import (
	"github.com/indigo-web/indigo/http"
	"io"
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
