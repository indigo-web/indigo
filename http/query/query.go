package query

import (
	"github.com/indigo-web/indigo/internal/keyvalue"
	"github.com/indigo-web/indigo/internal/qparams"
	"github.com/indigo-web/indigo/internal/urlencoded"
	"github.com/indigo-web/utils/uf"
)

// Params are parsed key-value pairs from the query itself
type Params = *keyvalue.Storage

// Query is an entity for lazy accessing the query
type Query struct {
	params Params
	raw    []byte
}

func New(params Params) Query {
	return Query{params: params}
}

// Cook parses the query and returns Params
//
// Recommendation: consider invoking this method only once, as repeatedly parsing big
// enough strings may be quite expensive
func (q *Query) Cook() (Params, error) {
	return q.params, qparams.Parse(q.raw, qparams.Into(q.params), urlencoded.Decode)
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

// Clear empties all the parsed parameters. Used mostly in internal purposes
func (q *Query) Clear() {
	q.raw = nil
	q.params.Clear()
}
