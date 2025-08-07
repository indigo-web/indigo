package codec

import (
	"io"

	"github.com/indigo-web/indigo/http"
	"github.com/klauspost/compress/gzip"
)

// TODO: pass this via parameters?
const decompressorBufferSize = 4096

var _ Codec = new(GZIP)

type GZIP struct{}

func NewGZIP() GZIP {
	return GZIP{}
}

func (GZIP) Token() string {
	return "gzip"
}

func (g GZIP) New() Instance {
	return newGZIPCodec(make([]byte, decompressorBufferSize))
}

var _ Instance = new(gzipCodec)

type gzipCodec struct {
	adapter *readerAdapter
	w       *gzip.Writer // compressor
	r       gzip.Reader  // decompressor
	buff    []byte
}

func newGZIPCodec(buff []byte) *gzipCodec {
	return &gzipCodec{
		adapter: newAdapter(),
		w:       gzip.NewWriter(nil),
		buff:    buff,
	}
}

func (g *gzipCodec) ResetCompressor(w io.Writer) {
	g.w.Reset(w)
}

func (g *gzipCodec) Write(p []byte) (n int, err error) {
	return g.w.Write(p)
}

func (g *gzipCodec) Flush() error {
	return g.w.Close()
}

func (g *gzipCodec) ResetDecompressor(source http.Fetcher) error {
	g.adapter.Reset(source)
	return g.r.Reset(g.adapter)
}

func (g *gzipCodec) Fetch() ([]byte, error) {
	n, err := g.r.Read(g.buff)
	return g.buff[:n], err
}
