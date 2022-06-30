package http

type Headers map[string][]byte

func (h Headers) Set(key string, value []byte) {
	oldValue, found := h[key]

	if !found || cap(oldValue) < len(value) {
		oldValue = make([]byte, len(value))
		h[key] = oldValue
	} else if len(oldValue) > len(value) {
		h[key] = oldValue[:len(value)]
	}

	copy(oldValue, value)
}
