package http

import (
	"io"
)

// Codec is a unified interface for any codecutil, which will later be checked for compatibility
// with Encoder and/or Decoder
type Codec interface {
	Tokens() []string
}

// CompressorFabric is used to instantiate compressors on demand. There exists at most
// one instance per connection.
type CompressorFabric interface {
	NewCompressor() Compressor
}

// DecompressorFabric is used to instantiate decompressors on demand. There exists at most
// one instance per connection.
type DecompressorFabric interface {
	NewDecompressor() Decompressor
}

type Compressor interface {
	io.ReadCloser
}

type Decompressor interface {
	Fetcher
	Reset(source Fetcher) error
}
