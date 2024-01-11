package stash

import "io"

// Source is a function that returns new data on demand
type Source func() ([]byte, error)

var _ io.Reader = &Reader{}

// Reader covers Source function in the manner, so it implements the io.Reader
type Reader struct {
	source  Source
	pending []byte
	error   error
}

func New() *Reader {
	return new(Reader)
}

func (r *Reader) Read(b []byte) (n int, err error) {
	if len(r.pending) == 0 && r.error == nil {
		r.refill()
	}

	n = copy(b, r.pending)
	r.pending = r.pending[n:]

	if len(r.pending) == 0 && r.error != nil {
		err = r.error
	}

	return n, err
}

func (r *Reader) refill() {
	r.pending, r.error = r.source()
}

func (r *Reader) Reset(src Source) {
	r.source = src
	r.pending = nil
	r.error = nil
}
