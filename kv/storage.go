package kv

import (
	"github.com/indigo-web/indigo/internal/strutil"
	"iter"
)

type Pair struct {
	Key, Value string
}

// TODO: add `deleted` attribute to pairs, so by that we can delete and update already existing
// TODO: entries

// Storage is an associative structure for storing (string, string) pairs. It acts as a map but
// uses linear search instead, which proves to be more efficient on relatively low amount of
// entries, which often enough is the case.
type Storage struct {
	pairs      []Pair
	uniqueBuff []string
	valuesBuff []string
}

func New() *Storage {
	return new(Storage)
}

// NewPrealloc returns an instance of Storage with pre-allocated underlying storage.
func NewPrealloc(n int) *Storage {
	return &Storage{
		pairs: make([]Pair, 0, n),
	}
}

// NewFromMap returns a new instance with already inserted values from given map.
// Note: as maps are unordered, resulting underlying structure will also contain unordered
// pairs.
func NewFromMap(m map[string][]string) *Storage {
	// this actually doesn't always allocate exactly enough sized slice, as we don't
	// count amount of _values_, only _keys_, where each key may contain more  (or less)
	// than 1 value. But this doesn't actually matter, as this job is made just once
	// per client, so considered not to be a hot path
	kv := NewPrealloc(len(m))

	for key, values := range m {
		for _, value := range values {
			kv.Add(key, value)
		}
	}

	return kv
}

// Add adds a new pair of key and value.
func (s *Storage) Add(key, value string) *Storage {
	s.pairs = append(s.pairs, Pair{
		Key:   key,
		Value: value,
	})
	return s
}

// Value returns the first value, corresponding to the key. Otherwise, empty string is returned
func (s *Storage) Value(key string) string {
	return s.ValueOr(key, "")
}

// ValueOr returns either the first value corresponding to the key or custom value, defined
// via the second parameter.
func (s *Storage) ValueOr(key, or string) string {
	value, found := s.Get(key)
	if !found {
		return or
	}

	return value
}

// Get returns a value and a bool, indicating whether the value was found. If it wasn't, it'll
// be an empty string.
func (s *Storage) Get(key string) (value string, found bool) {
	for _, pair := range s.pairs {
		if strutil.CmpFold(key, pair.Key) {
			return pair.Value, true
		}
	}

	return "", false
}

// Values returns all values by the key. Returns nil if key doesn't exist.
//
// WARNING: calling it twice will override values, returned by the first call. Consider
// copying the returned slice for safe use.
func (s *Storage) Values(key string) (values []string) {
	s.valuesBuff = s.valuesBuff[:0]

	for _, pair := range s.pairs {
		if strutil.CmpFold(pair.Key, key) {
			s.valuesBuff = append(s.valuesBuff, pair.Value)
		}
	}

	if len(s.valuesBuff) == 0 {
		return nil
	}

	return s.valuesBuff
}

// Keys returns all unique presented keys.
//
// WARNING: calling it twice will override values, returned by the first call. Consider
// copying the returned slice for safe use.
func (s *Storage) Keys() []string {
	s.uniqueBuff = s.uniqueBuff[:0]

	for _, pair := range s.pairs {
		if contains(s.uniqueBuff, pair.Key) {
			continue
		}

		s.uniqueBuff = append(s.uniqueBuff, pair.Key)
	}

	return s.uniqueBuff
}

// Iter returns an iterator over the pairs.
func (s *Storage) Iter() iter.Seq2[string, string] {
	return func(yield func(string, string) bool) {
		for _, pair := range s.pairs {
			if !yield(pair.Key, pair.Value) {
				break
			}
		}

		return
	}
}

// Has indicates, whether there's an entry of the key.
func (s *Storage) Has(key string) bool {
	for _, pair := range s.pairs {
		if strutil.CmpFold(key, pair.Key) {
			return true
		}
	}

	return false
}

// Len returns a number of stored pairs.
func (s *Storage) Len() int {
	return len(s.pairs)
}

func (s *Storage) Empty() bool {
	return s.Len() == 0
}

// Clone creates a deep copy, which may be used later or stored somewhere safely. However,
// it comes at cost of multiple allocations.
func (s *Storage) Clone() *Storage {
	return &Storage{
		pairs:      clone(s.pairs),
		uniqueBuff: clone(s.uniqueBuff),
		valuesBuff: clone(s.valuesBuff),
	}
}

// Expose exposes the underlying pairs slice.
func (s *Storage) Expose() []Pair {
	return s.pairs
}

// Clear all the entries. However, all the allocated space won't be freed.
func (s *Storage) Clear() *Storage {
	s.pairs = s.pairs[:0]
	return s
}

func (s *Storage) ensureNotNil(buff []string) []string {
	if buff == nil {
		buff = make([]string, 0, len(s.pairs))
	}

	return buff
}

func contains(collection []string, key string) bool {
	for _, element := range collection {
		if strutil.CmpFold(element, key) {
			return true
		}
	}

	return false
}

func clone[T any](source []T) []T {
	if len(source) == 0 {
		return nil
	}

	newSlice := make([]T, len(source))
	copy(newSlice, source)

	return newSlice
}
