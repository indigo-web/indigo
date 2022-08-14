package url

type (
	rawQuery    []byte
	parsedQuery map[string][]byte
)

// Query struct is simply url parameters parser. I have a choice to implement it
// in to ways:
// 1) lazy parsing - if we wanna get a specific query, but it is not in parsedParams,
// then we simply start parsing rawQuery until key we need will be met
// 2) naive parsing - as we have a limited length of URL, we may do not mind about
// flood and just parse everything until the end
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

func (q *Query) Get(key string) []byte {
	panic("Implement me!")
}
