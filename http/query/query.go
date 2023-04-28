package query

import (
	"errors"

	"github.com/indigo-web/indigo/internal/queryparser"
)

var ErrNoSuchKey = errors.New("desired key does not exists")

type (
	rawQuery   []byte
	Map        = map[string]string
	mapFactory func() Map
)

// Query is optional, it may contain rawQuery, but it will not be parsed until
// needed
type Query struct {
	parsedQuery  Map
	queryFactory mapFactory
	rawQuery     rawQuery
}

func NewQuery(queryFactory mapFactory) Query {
	return Query{
		queryFactory: queryFactory,
	}
}

// Set is responsible for setting a raw value of query. Each call
// resets parsedQuery value to nil (query bytearray must be parsed
// again)
func (q *Query) Set(raw []byte) {
	q.rawQuery = raw
	q.parsedQuery = nil
}

// Get is responsible for getting a key from query. In case this
// method is called a first time since rawQuery was set (or not set
// at all), rawQuery bytearray will be parsed and value returned
// (or ErrNoSuchKey instead). In case of invalid query bytearray,
// ErrBadQuery will be returned
func (q *Query) Get(key string) (value string, err error) {
	if q.parsedQuery == nil {
		q.parsedQuery, err = queryparser.Parse(q.rawQuery, q.queryFactory)
		if err != nil {
			return "", err
		}
	}

	value, found := q.parsedQuery[key]
	if !found {
		err = ErrNoSuchKey
	}

	return value, err
}

// Raw just returns a raw value of query as it is
func (q *Query) Raw() []byte {
	return q.rawQuery
}
