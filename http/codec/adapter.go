package codec

type readerAdapter struct {
	retriever Retriever
	err       error
	data      []byte
}

func (r *readerAdapter) Read(b []byte) (n int, err error) {
	if len(r.data) == 0 {
		if r.err != nil {
			return 0, r.err
		}

		r.data, r.err = r.retriever.Retrieve()
	}

	n = copy(b, r.data)
	r.data = r.data[n:]
	if len(r.data) == 0 {
		err = r.err
	}

	return n, err
}

func (r *readerAdapter) Reset(retriever Retriever) {
	*r = readerAdapter{retriever: retriever}
}
