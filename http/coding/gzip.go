package coding

import (
	"github.com/indigo-web/indigo/http"
	"github.com/klauspost/compress/gzip"
)

type GZIP struct {
	r    gzip.Reader
	buff []byte
}

func NewGZIP(buffsize int) *GZIP {
	return &GZIP{
		r:    gzip.Reader{},
		buff: make([]byte, buffsize),
	}
}

func (g *GZIP) Retrieve() ([]byte, error) {
	n, err := g.r.Read(g.buff)
	return g.buff[:n], err
}

func (g *GZIP) Reset(source http.Retriever) error {
	return g.r.Reset(nil)
}
