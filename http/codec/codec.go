package codec

import (
	"io"
)

// Codec is a unified interface for any codecutil, which will later be checked for compatibility
// with Encoder and/or Decoder
type Codec interface {
	Tokens() []string
}

// Encoder is a fabric of compressors. A new compressor is instantiated per connection and
// lazily, i.e. on demand.
type Encoder interface {
	NewCompressor() Compressor
}

// Decoder is a fabric of decompressors. A new decompressor is instantiated per connection and
// lazily, i.e. on demand.
type Decoder interface {
	NewDecompressor() Decompressor
}

type Compressor interface {
	io.ReadCloser
}

// Retriever is re-defined from package http in order to avoid the cyclic import, which
// occurs otherwise.
type Retriever interface {
	Retrieve() ([]byte, error)
}

type Decompressor interface {
	Retrieve() ([]byte, error)
	Reset(source Retriever) error
}
