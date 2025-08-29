package codec

import (
	"io"

	"github.com/indigo-web/indigo/http"
)

var _ Codec = baseCodec{}

type instantiator = func() Instance

type baseCodec struct {
	token   string
	newInst instantiator
}

func newBaseCodec(token string, newInst instantiator) baseCodec {
	return baseCodec{
		token:   token,
		newInst: newInst,
	}
}

func (b baseCodec) Token() string {
	return b.token
}

func (b baseCodec) New() Instance {
	return b.newInst()
}

var _ Instance = new(baseInstance)

type (
	decoderResetter = func(io.Reader, *readerAdapter) error

	writeResetter interface {
		io.WriteCloser
		Reset(dst io.Writer)
	}
)

type baseInstance struct {
	reset   decoderResetter
	adapter *readerAdapter
	w       writeResetter // compressor
	r       io.Reader     // decompressor
	dst     io.Closer
	buff    []byte
}

func newBaseInstance(encoder writeResetter, decoder io.Reader, reset decoderResetter) instantiator {
	return func() Instance {
		return &baseInstance{
			reset:   reset,
			adapter: newAdapter(),
			w:       encoder,
			r:       decoder,
		}
	}
}

func (b *baseInstance) ResetCompressor(w io.Writer) {
	b.w.Reset(w)
	b.dst = nil

	if c, ok := w.(io.Closer); ok {
		b.dst = c
	}
}

func (b *baseInstance) Write(p []byte) (n int, err error) {
	return b.w.Write(p)
}

func (b *baseInstance) Close() error {
	if err := b.w.Close(); err != nil {
		return err
	}

	if b.dst != nil {
		return b.dst.Close()
	}

	return nil
}

func (b *baseInstance) ResetDecompressor(source http.Fetcher, bufferSize int) error {
	if cap(b.buff) < bufferSize {
		b.buff = make([]byte, bufferSize)
	}

	b.adapter.Reset(source)

	return b.reset(b.r, b.adapter)
}

func (b *baseInstance) Fetch() ([]byte, error) {
	n, err := b.r.Read(b.buff)
	return b.buff[:n], err
}

func genericResetter(r io.Reader, adapter *readerAdapter) error {
	type resetter interface {
		Reset(r io.Reader) error
	}

	if reset, ok := r.(resetter); ok {
		return reset.Reset(adapter)
	}

	return nil
}

type readerAdapter struct {
	fetcher http.Fetcher
	err     error
	data    []byte
}

func newAdapter() *readerAdapter {
	return new(readerAdapter)
}

func (r *readerAdapter) Read(b []byte) (n int, err error) {
	if len(r.data) == 0 {
		if r.err != nil {
			return 0, r.err
		}

		r.data, r.err = r.fetcher.Fetch()
	}

	n = copy(b, r.data)
	r.data = r.data[n:]
	if len(r.data) == 0 {
		err = r.err
	}

	return n, err
}

func (r *readerAdapter) Reset(fetcher http.Fetcher) {
	*r = readerAdapter{fetcher: fetcher}
}
