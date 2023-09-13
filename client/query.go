package client

type Query map[string][]string

func NewQuery() Query {
	return make(Query)
}

func (q Query) WithValue(key string, values ...string) {
	q[key] = append(q[key], values...)
}
