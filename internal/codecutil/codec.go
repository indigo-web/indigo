package codecutil

import (
	"iter"

	"github.com/indigo-web/indigo/http/codec"
	"github.com/indigo-web/indigo/internal/strutil"
)

type Cache struct {
	codecs    []codec.Codec
	instances []codec.Instance
}

func NewCache(codecs []codec.Codec) Cache {
	return Cache{
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

func (c Cache) AcceptEncodings() string {
	// TODO: somehow cache the value so we don't have to build the string every time
	// TODO: a new connection establishes?

	if len(c.codecs) == 0 {
		return "identity"
	}

	return strutil.Join(traverseTokens(c.codecs), ",")
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
