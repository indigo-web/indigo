package codec

import (
	"github.com/indigo-web/indigo/http"
	"github.com/klauspost/compress/gzip"
)

type GZIP struct {
	adapter *readerAdapter
	buff    []byte
	r       gzip.Reader
}

func NewGZIP(buffsize int) *GZIP {
	return &GZIP{
		adapter: new(readerAdapter),
		buff:    make([]byte, buffsize),
	}
}

func (g *GZIP) Fetch() ([]byte, error) {
	n, err := g.r.Read(g.buff)
	return g.buff[:n], err
}

func (g *GZIP) Reset(source http.Fetcher) error {
	g.adapter.Reset(source)
	return g.r.Reset(g.adapter)
}
