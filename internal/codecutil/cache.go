package codecutil

import (
	"strings"

	"github.com/indigo-web/indigo/http/codec"
)

type Cache struct {
	accept    string
	codecs    []codec.Codec
	instances []codec.Instance
}

func NewCache(codecs []codec.Codec, acceptString string) Cache {
	return Cache{
		accept:    acceptString,
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

func AcceptEncoding(codecs []codec.Codec) string {
	if len(codecs) == 0 {
		return "identity"
	}

	var b strings.Builder

	b.WriteString(codecs[0].Token())
	for _, c := range codecs[1:] {
		b.WriteString(", ")
		b.WriteString(c.Token())
	}

	return b.String()
}
