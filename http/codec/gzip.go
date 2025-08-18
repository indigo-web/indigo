package codec

import (
	"io"

	"github.com/indigo-web/indigo/http"
	"github.com/klauspost/compress/gzip"
)

// TODO: pass this via parameters?
const gzipBufferSize = 4096

var _ Codec = new(GZIP)

type GZIP struct{}

func NewGZIP() GZIP {
	return GZIP{}
}

func (GZIP) Token() string {
	return "gzip"
}

func (g GZIP) New() Instance {
	return newGZIPCodec(make([]byte, gzipBufferSize))
}

var _ Instance = new(gzipCodec)

type gzipCodec struct {
	adapter *readerAdapter
	w       *gzip.Writer // compressor
	r       gzip.Reader  // decompressor
	wout    io.Closer
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

	if c, ok := w.(io.Closer); ok {
		g.wout = c
	}
}

func (g *gzipCodec) Write(p []byte) (n int, err error) {
	// TODO: the compressor spams with Write() calls. This will cause significant performance downgrade,
	// TODO: as each individual Write() call results in transferring the passed data over the network.
	// TODO: Buffer this somewhere to at least 4096 (by default). Make the behaviour disable-able.
	return g.w.Write(p)
}

func (g *gzipCodec) Close() error {
	if err := g.w.Close(); err != nil {
		return err
	}

	if g.wout != nil {
		return g.wout.Close()
	}

	return nil
}

func (g *gzipCodec) ResetDecompressor(source http.Fetcher) error {
	g.adapter.Reset(source)
	return g.r.Reset(g.adapter)
}

func (g *gzipCodec) Fetch() ([]byte, error) {
	n, err := g.r.Read(g.buff)
	return g.buff[:n], err
}
