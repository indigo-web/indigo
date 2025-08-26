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
	io.WriteCloser
	ResetCompressor(w io.Writer)
}

type Decompressor interface {
	http.Fetcher
	ResetDecompressor(source http.Fetcher, bufferSize int) error
}
