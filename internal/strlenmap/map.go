package strlenmap

type Map[T any] struct {
	buckets []bucket[T]
	Values  []T
}

func New[T any]() *Map[T] {
	return &Map[T]{}
}

func (m *Map[T]) Get(key string) (v T, found bool) {
	if len(key) >= len(m.buckets) {
		return v, false
	}

	for _, entry := range m.buckets[len(key)] {
		if entry.Key == key {
			return entry.Value, true
		}
	}

	return v, false
}

// Insert inserts a value. If the key already exists, it won't be overridden. However,
// a new value will be ignored, and the first one added will always be used
func (m *Map[T]) Insert(key string, value T) {
	if len(m.buckets) <= len(key) {
		m.resize(len(key) + 1)
	}

	m.buckets[len(key)] = append(m.buckets[len(key)], buckEntry[T]{
		Key:   key,
		Value: value,
	})
}

func (m *Map[T]) resize(n int) {
	newBuckets := make([]bucket[T], n)
	copy(newBuckets, m.buckets)
	m.buckets = newBuckets
}

type bucket[T any] []buckEntry[T]

type buckEntry[T any] struct {
	Key   string
	Value T
}
