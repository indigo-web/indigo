package internal

func Keys[K comparable, V any](from map[K]V) []K {
	keys := make([]K, 0, len(from))
	for key := range from {
		keys = append(keys, key)
	}

	return keys
}
