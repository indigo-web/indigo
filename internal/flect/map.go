package flect

type attrsMap struct {
	buckets []attrsMapBucket
}

func (a *attrsMap) Lookup(key string) (field fieldData, found bool) {
	if len(key) > len(a.buckets) {
		return fieldData{}, false
	}

	for _, entry := range a.buckets[len(key)] {
		if entry.Key == key {
			return entry.Value, true
		}
	}

	return fieldData{}, false
}

func (a *attrsMap) Insert(key string, value fieldData) {
	if len(a.buckets) < len(key) {
		a.grow(len(key))
	}

	a.buckets[len(key)] = append(a.buckets[len(key)], attrsMapEntry{
		Key:   key,
		Value: value,
	})
}

func (a *attrsMap) grow(n int) {
	newBuckets := make([]attrsMapBucket, n+1)
	copy(newBuckets, a.buckets)
	a.buckets = newBuckets
}

type attrsMapBucket []attrsMapEntry

type attrsMapEntry struct {
	Key   string
	Value fieldData
}
