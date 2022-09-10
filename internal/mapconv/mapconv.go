package mapconv

// Keys is currently used only in http/encodings/contentencodings.go, but
// this or other functions that may be added later may be used somewhere
// else
func Keys[K comparable, V any](from map[K]V) []K {
	keys := make([]K, 0, len(from))
	for key := range from {
		keys = append(keys, key)
	}

	return keys
}
