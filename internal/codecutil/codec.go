package codecutil

import (
	"iter"

	"github.com/indigo-web/indigo/http/codec"
	"github.com/indigo-web/indigo/internal/strutil"
)

type Cache struct {
	accept    string
	codecs    []codec.Codec
	instances []codec.Instance
}

func NewCache(codecs []codec.Codec) Cache {
	return Cache{
		// TODO: we're still allocating a string on every connection. Which we actually can avoid.
		accept:    acceptEncodings(codecs),
		codecs:    codecs,
		instances: make([]codec.Instance, len(codecs)),
	}
}

func (c Cache) find(token string) (int, codec.Codec) {
	for i, entry := range c.codecs {
		if entry.Token() == token {
			return i, entry
		}
	}

	return -1, nil
}

func (c Cache) Get(token string) codec.Instance {
	idx, cd := c.find(token)
	if idx == -1 {
		return nil
	}

	inst := c.instances[idx]
	if inst == nil {
		inst = cd.New()
		c.instances[idx] = inst
	}

	return inst
}

func (c Cache) AcceptEncoding() string {
	return c.accept
}

func acceptEncodings(codecs []codec.Codec) string {
	if len(codecs) == 0 {
		return "identity"
	}

	return strutil.Join(traverseTokens(codecs), ", ")
}

func traverseTokens(codecs []codec.Codec) iter.Seq[string] {
	return func(yield func(string) bool) {
		for _, c := range codecs {
			if !yield(c.Token()) {
				break
			}
		}
	}
}
