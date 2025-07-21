package codec

import (
	"github.com/indigo-web/indigo/http"
)

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
