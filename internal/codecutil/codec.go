package codecutil

import (
	"github.com/indigo-web/indigo/http"
)

type pair[A, B any] struct {
	A A
	B B
}

type constructor[T any] pair[string, func() T]

type (
	Compressors   = []constructor[http.Compressor]
	Decompressors = []constructor[http.Decompressor]
)

func Scatter(codecs []http.Codec) (cs Compressors, dcs Decompressors) {
	cs = make(Compressors, 0, len(codecs))
	dcs = make(Decompressors, 0, len(codecs))

	for _, c := range codecs {
		tokens := c.Tokens()

		if compressorFabric, ok := c.(http.CompressorFabric); ok {
			cs = appendCodec(cs, tokens, compressorFabric.NewCompressor)
		}

		if decompressorFabric, ok := c.(http.DecompressorFabric); ok {
			dcs = appendCodec(dcs, tokens, decompressorFabric.NewDecompressor)
		}
	}

	return cs, dcs
}

func appendCodec[T any](dst []constructor[T], tokens []string, constr func() T) []constructor[T] {
	for _, token := range tokens {
		dst = append(dst, constructor[T]{token, constr})
	}

	return dst
}

// Cache manages codecs by returning either one from cache or newly instantiated.
type Cache[T any] struct {
	constructors []constructor[T]
	cache        []pair[bool, T]
}

func NewCache[T any](constructors []constructor[T]) Cache[T] {
	return Cache[T]{
		constructors: constructors,
		cache:        make([]pair[bool, T], len(constructors)),
	}
}

func (c Cache[T]) Get(token string) (instance T, found bool) {
	for i, constr := range c.constructors {
		if constr.A == token {
			return c.get(i), true
		}
	}

	return instance, false
}

func (c Cache[T]) get(index int) T {
	cached := c.cache[index]
	if !cached.A {
		return c.constructors[index].B()
	}

	return cached.B
}
