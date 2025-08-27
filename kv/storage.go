package kv

import (
	"iter"
	"slices"

	"github.com/indigo-web/indigo/internal/strutil"
)

type Pair struct {
	Key, Value string
}

// Storage is an associative structure for storing (string, string) pairs. It acts as a map but
// uses linear search instead, which proves to be more efficient on relatively low amount of
// entries, which often enough is the case.
type Storage struct {
	deleted    int
	pairs      []Pair
	uniqueKeys []string
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
	kv := NewPrealloc(len(m))

	for key, values := range m {
		for _, value := range values {
			kv.Add(key, value)
		}
	}

	return kv
}

// NewFromPairs returns a new instance backed by the passed pairs as-is, without copying it.
// Use this method with care.
func NewFromPairs(pairs []Pair) *Storage {
	return &Storage{pairs: pairs}
}

// Add adds a new pair of key and value.
func (s *Storage) Add(key, value string) *Storage {
	if s.deleted == 0 {
		s.pairs = append(s.pairs, Pair{key, value})
		return s
	}

	for i, pair := range s.pairs {
		if len(pair.Key) == 0 {
			s.pairs[i] = Pair{key, value}
			break
		}
	}

	return s
}

// Set removes all the entries corresponding the key and sets the new value.
func (s *Storage) Set(key, value string) *Storage {
	freeIdx := s.delete(key)
	if freeIdx == -1 {
		return s.Add(key, value)
	}

	s.pairs[freeIdx] = Pair{key, value}
	s.deleted-- // as one of the deleted entries we've just overwritten
	return s
}

// Delete quasi removes all the entries corresponding the key. In fact, an empty key is assigned
// to each eligible pair.
func (s *Storage) Delete(key string) *Storage {
	s.delete(key)
	return s
}

func (s *Storage) delete(key string) (firstDeleted int) {
	firstDeleted = -1

	for i, pair := range s.pairs {
		if strutil.CmpFoldFast(pair.Key, key) {
			s.pairs[i].Key = ""
			s.deleted++

			if firstDeleted == -1 {
				firstDeleted = i
			}
		}
	}

	return firstDeleted
}

// Value returns the first value, corresponding to the key. Otherwise, empty string is returned
func (s *Storage) Value(key string) string {
	return s.ValueOr(key, "")
}

// ValueOr returns either the first found value or the second parameter.
func (s *Storage) ValueOr(key, otherwise string) string {
	if value, found := s.Lookup(key); found {
		return value
	}

	return otherwise
}

// Lookup returns the first found value and indicates the success via a bool flag.
func (s *Storage) Lookup(key string) (value string, found bool) {
	if i := s.findNext(key, 0); i != -1 {
		return s.pairs[i].Value, true
	}

	return "", false
}

// Values returns an iterator over all the values corresponding the given key.
func (s *Storage) Values(key string) iter.Seq[string] {
	return func(yield func(string) bool) {
		for _, pair := range s.pairs {
			if strutil.CmpFoldFast(pair.Key, key) {
				if !yield(pair.Value) {
					break
				}
			}
		}
	}
}

// Keys returns an iterator over all the unique non-normalized keys.
func (s *Storage) Keys() iter.Seq[string] {
	return func(yield func(string) bool) {
		s.uniqueKeys = s.uniqueKeys[:0]

		for key := range s.Pairs() {
			if contains(s.uniqueKeys, key) {
				continue
			}

			if !yield(key) {
				return
			}

			s.uniqueKeys = append(s.uniqueKeys, key)
		}
	}
}

// Pairs return an iterator over all the header field pairs presented. Multiple values are
// represented by occupying multiple pairs sharing the same key. Array values can occupy
// non-consecutive pairs.
func (s *Storage) Pairs() iter.Seq2[string, string] {
	return func(yield func(string, string) bool) {
		for _, pair := range s.pairs {
			if len(pair.Key) == 0 {
				continue
			}

			if !yield(pair.Key, pair.Value) {
				return
			}
		}
	}
}

// Has indicates, whether there's an entry of the key.
func (s *Storage) Has(key string) bool {
	_, found := s.Lookup(key)
	return found
}

// Len returns a number of stored pairs. Duplicate key entries are counted in.
func (s *Storage) Len() int {
	return len(s.pairs) - s.deleted
}

func (s *Storage) Empty() bool {
	return s.Len() == 0
}

// Clone creates a deep copy, which may be used later or stored somewhere safely. However,
// it comes at cost of multiple allocations.
func (s *Storage) Clone() *Storage {
	return &Storage{
		pairs: slices.Clone(s.pairs),
	}
}

// Expose exposes the underlying pairs slice. Unlike Pairs, it also contains all the deleted
// empty-key pairs
func (s *Storage) Expose() []Pair {
	return s.pairs
}

// Clear all the entries. However, all the allocated space won't be freed.
func (s *Storage) Clear() *Storage {
	s.pairs = s.pairs[:0]
	s.deleted = 0
	return s
}

func (s *Storage) findNext(key string, startAt int) int {
	for i := startAt; i < len(s.pairs); i++ {
		if strutil.CmpFoldFast(s.pairs[i].Key, key) {
			return i
		}
	}

	return -1
}

func contains(collection []string, key string) bool {
	for _, element := range collection {
		if strutil.CmpFoldFast(element, key) {
			return true
		}
	}

	return false
}
