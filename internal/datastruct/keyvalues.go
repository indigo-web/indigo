package datastruct

import (
	"github.com/indigo-web/iter"
	"github.com/indigo-web/utils/strcomp"
)

type Pair struct {
	Key, Value string
}

// KeyValue is a generic structure for storing pairs of string-string. It is used across
// the whole database. For example, it is primarily used for request headers, however
// used as well as a storage for URI query, dynamic routing parameters, etc.
type KeyValue struct {
	pairs      []Pair
	uniqueBuff []string
	valuesBuff []string
}

// NewKeyValueFromMap returns a new instance with already inserted values from given map.
// Note: as maps are unordered, resulting underlying structure will also contain unordered
// pairs
func NewKeyValueFromMap(m map[string][]string) *KeyValue {
	// this actually doesn't always allocate exactly enough sized slice, as we don't
	// count amount of _values_, only _keys_, where each key may contain more  (or less)
	// than 1 value. But this doesn't actually matter, as this job is made just once
	// per client, so considered not to be a hot path
	kv := NewKeyValuePreAlloc(len(m))

	for key, values := range m {
		for _, value := range values {
			kv.Add(key, value)
		}
	}

	return kv
}

// NewKeyValuePreAlloc returns an instance of KeyValue with pre-allocated underlying storage
func NewKeyValuePreAlloc(n int) *KeyValue {
	return &KeyValue{
		pairs: make([]Pair, 0, n),
	}
}

func NewKeyValue() *KeyValue {
	return NewKeyValuePreAlloc(0)
}

// Add adds a new pair of key and value
func (k *KeyValue) Add(key, value string) *KeyValue {
	k.pairs = append(k.pairs, Pair{
		Key:   key,
		Value: value,
	})
	return k
}

// Value returns the first value, corresponding to the key. Otherwise, empty string is returned
func (k *KeyValue) Value(key string) string {
	return k.ValueOr(key, "")
}

// ValueOr returns either the first value corresponding to the key or custom value, defined
// via the second parameter
func (k *KeyValue) ValueOr(key, or string) string {
	value, found := k.Get(key)
	if !found {
		return or
	}

	return value
}

// Get returns a value corresponding to the key and a bool, indicating whether the key
// exists. In case it doesn't, empty string will be returned either
func (k *KeyValue) Get(key string) (string, bool) {
	for _, pair := range k.pairs {
		if strcomp.EqualFold(key, pair.Key) {
			return pair.Value, true
		}
	}

	return "", false
}

// Values returns all values by the key. Returns nil if key doesn't exist.
//
// WARNING: calling it twice will override values, returned by the first call. Consider
// copying the returned slice for safe use
func (k *KeyValue) Values(key string) (values []string) {
	k.valuesBuff = k.valuesBuff[:0]

	for _, pair := range k.pairs {
		if strcomp.EqualFold(pair.Key, key) {
			k.valuesBuff = append(k.valuesBuff, pair.Value)
		}
	}

	if len(k.valuesBuff) == 0 {
		return nil
	}

	return k.valuesBuff
}

// Keys returns all unique presented keys.
//
// WARNING: calling it twice will override values, returned by the first call. Consider
// copying the returned slice for safe use
func (k *KeyValue) Keys() []string {
	k.uniqueBuff = k.uniqueBuff[:0]

	for _, pair := range k.pairs {
		if contains(k.uniqueBuff, pair.Key) {
			continue
		}

		k.uniqueBuff = append(k.uniqueBuff, pair.Key)
	}

	return k.uniqueBuff
}

// Iter returns an iterator over the pairs
func (k *KeyValue) Iter() iter.Iterator[Pair] {
	return iter.Slice(k.pairs)
}

// Has indicates, whether there's an entry of the key
func (k *KeyValue) Has(key string) bool {
	for _, pair := range k.pairs {
		if strcomp.EqualFold(key, pair.Key) {
			return true
		}
	}

	return false
}

// Clone creates a deep copy, which may be used later or stored somewhere safely. However,
// it comes at cost of multiple allocations
func (k *KeyValue) Clone() *KeyValue {
	return &KeyValue{
		pairs:      clone(k.pairs),
		uniqueBuff: clone(k.uniqueBuff),
		valuesBuff: clone(k.valuesBuff),
	}
}

// Unwrap reveals underlying data structure. Try to avoid the method if possible, as
// changing the signature may not affect a major version
func (k *KeyValue) Unwrap() []Pair {
	return k.pairs
}

// Clear all the entries. However, all the allocated space won't be freed
func (k *KeyValue) Clear() {
	k.pairs = k.pairs[:0]
}

func (k *KeyValue) ensureNotNil(buff []string) []string {
	if buff == nil {
		buff = make([]string, 0, len(k.pairs))
	}

	return buff
}

func contains(collection []string, key string) bool {
	for _, element := range collection {
		if strcomp.EqualFold(element, key) {
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
