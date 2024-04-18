package query

import (
	"errors"
	"github.com/indigo-web/indigo/http/query/internal"
	"github.com/indigo-web/indigo/internal/keyvalue"
)

var ErrNoSuchKey = errors.New("no entry by the key")

type Params = *keyvalue.Storage

// Query is a lazy structure for accessing URI parameters. Its laziness is defined
// by the fact that parameters won't be parsed until requested
type Query struct {
	parsed bool
	params Params
	raw    []byte
}

func NewQuery(underlying *keyvalue.Storage) *Query {
	return &Query{
		params: underlying,
	}
}

// Set is responsible for setting a raw value of query. Each call
// resets parsedQuery value to nil (query bytearray must be parsed
// again)
func (q *Query) Set(raw []byte) {
	q.raw = raw

	if q.parsed {
		q.parsed = false
		q.params.Clear()
	}
}

// Get is responsible for getting a key from query. In case this
// method is called a first time since rawQuery was set (or not set
// at all), rawQuery bytearray will be parsed and value returned
// (or ErrNoSuchKey instead). In case of invalid query bytearray,
// ErrBadQuery will be returned
func (q *Query) Get(key string) (value string, err error) {
	if err = q.parse(); err != nil {
		return "", err
	}

	value, found := q.params.Get(key)
	if !found {
		err = ErrNoSuchKey
	}

	return value, err
}

// Unwrap returns all query parameters. If error occurred, nil parameters will be returned
func (q *Query) Unwrap() (Params, error) {
	return q.params, q.parse()
}

// Raw just returns a raw value of query as it is
func (q *Query) Raw() []byte {
	return q.raw
}

func (q *Query) parse() error {
	if q.parsed {
		return nil
	}

	q.parsed = true

	return internal.Parse(q.raw, q.params)
}
