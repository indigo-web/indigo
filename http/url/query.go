package url

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

func NewQuery(raw []byte) Query {
	return Query{
		parsedQuery: make(parsedQuery),
		rawQuery:    raw,
	}
}

func (q *Query) Set(raw []byte) {
	q.rawQuery = append(q.rawQuery[:0], raw...)
}

func (q *Query) Get(key string) (value []byte, found bool) {
	if q.parsedQuery == nil {
		q.parsedQuery = parse(q.rawQuery)
	}

	value, found = q.parsedQuery[key]
	return value, found
}

func parse(raw []byte) parsedQuery {
	panic("implement me!")
}
