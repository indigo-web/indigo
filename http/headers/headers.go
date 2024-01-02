package headers

import "github.com/indigo-web/indigo/internal/datastruct"

type (
	Header  = datastruct.Pair
	Headers = datastruct.KeyValue
)

func NewPrealloc(n int) *Headers {
	return datastruct.NewKeyValuePreAlloc(n)
}

func New() *Headers {
	return NewPrealloc(0)
}

func NewFromMap(m map[string][]string) *Headers {
	return datastruct.NewKeyValueFromMap(m)
}
