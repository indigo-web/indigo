package url

import (
	"indigo/errors"
	"indigo/http/url/queryparser"
)

type (
	rawQuery    []byte
	parsedQuery map[string][]byte
)

// Query is optional, it may contain rawQuery, but it will not be parsed until
// needed
type Query struct {
	parsedQuery parsedQuery
	rawQuery    rawQuery
}

func NewQuery(buff []byte) Query {
	return Query{
		rawQuery: buff,
	}
}

func (q *Query) Set(raw []byte) {
	// TODO: add to settings a new setting of initial parsedQuery capacity
	//       and maximal number of query key-values allowed
	q.parsedQuery = nil
	q.rawQuery = append(q.rawQuery[:0], raw...)
}

func (q *Query) Get(key string) (value []byte, err error) {
	if q.parsedQuery == nil {
		q.parsedQuery, err = queryparser.Parse(q.rawQuery)
		if err != nil {
			return nil, err
		}
	}

	value, found := q.parsedQuery[key]
	if !found {
		err = errors.ErrNoSuchKey
	}

	return value, err
}
