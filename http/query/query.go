package query

import (
	"github.com/indigo-web/indigo/config"
	"github.com/indigo-web/indigo/internal/keyvalue"
	"github.com/indigo-web/indigo/internal/qparams"
	"github.com/indigo-web/indigo/internal/urlencoded"
	"github.com/indigo-web/utils/uf"
)

// Params are parsed key-value pairs from the query itself
type Params = *keyvalue.Storage

// Query is an entity for lazy accessing the query
type Query struct {
	cfg       *config.Config
	params    Params
	raw, buff []byte
}

func New(params Params, cfg *config.Config) Query {
	return Query{
		cfg:    cfg,
		params: params,
	}
}

// Cook parses the query and returns it in as Params.
//
// Note: this method can be quite expensive. Consider saving and re-using the result.
func (q *Query) Cook() (params Params, err error) {
	if q.buff == nil {
		q.buff = make([]byte, q.cfg.URL.Query.BufferPrealloc)
	}

	defFlagValue := q.cfg.URL.Query.DefaultFlagValue
	q.buff, err = qparams.Parse(q.raw, q.buff[:0], qparams.Into(q.params), urlencoded.ExtendedDecode, defFlagValue)
	return q.params, err
}

// Bytes returns the actual query, as it has been received
func (q *Query) Bytes() []byte {
	return q.raw
}

// String returns the actual query, as it has been received
//
// Note: returned string is unsafe and must never be used after the request has been
// processed
func (q *Query) String() string {
	return uf.B2S(q.raw)
}

// Update sets the new raw query string. This doesn't cause the existing parameters
// to be parsed again, however. Used mostly in internal purposes
func (q *Query) Update(new []byte) {
	q.raw = new
}

// Reset empties all the parsed parameters. Used mostly in internal purposes
func (q *Query) Reset() {
	q.raw = nil
	q.params.Clear()
}
