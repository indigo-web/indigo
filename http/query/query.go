package query

import (
	"errors"
	"github.com/indigo-web/indigo/http/headers"

	"github.com/indigo-web/indigo/internal/queryparser"
)

var ErrNoSuchKey = errors.New("no such key")

// Query is optional, it may contain rawQuery, but it will not be parsed until
// needed
type Query struct {
	parsed bool
	query  *headers.Headers
	raw    []byte
}

func NewQuery(query *headers.Headers) Query {
	return Query{
		query: query,
	}
}

// Set is responsible for setting a raw value of query. Each call
// resets parsedQuery value to nil (query bytearray must be parsed
// again)
func (q *Query) Set(raw []byte) {
	q.raw = raw

	if q.parsed {
		q.query.Clear()
	}

	q.parsed = false
}

// Get is responsible for getting a key from query. In case this
// method is called a first time since rawQuery was set (or not set
// at all), rawQuery bytearray will be parsed and value returned
// (or ErrNoSuchKey instead). In case of invalid query bytearray,
// ErrBadQuery will be returned
func (q *Query) Get(key string) (value string, err error) {
	if !q.parsed {
		err = queryparser.Parse(q.raw, q.query)
		if err != nil {
			return "", err
		}

		q.parsed = true
	}

	value, found := q.query.Get(key)
	if !found {
		err = ErrNoSuchKey
	}

	return value, err
}

// Raw just returns a raw value of query as it is
func (q *Query) Raw() []byte {
	return q.raw
}
