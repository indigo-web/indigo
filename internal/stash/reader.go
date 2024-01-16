package stash

type Retriever func() ([]byte, error)

// Reader covers Source function in the manner, so it implements the io.Reader
type Reader struct {
	source  Retriever
	pending []byte
	error   error
}

func New(source Retriever) *Reader {
	return &Reader{
		source: source,
	}
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

func (r *Reader) Reset() {
	r.pending = nil
	r.error = nil
}
