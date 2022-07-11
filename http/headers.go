package http

import (
	"indigo/internal"
)

type Headers map[string][]byte

func (h Headers) Set(key []byte, value []byte) {
	oldValue, found := h[internal.B2S(key)]

	if !found || cap(oldValue) < len(value) {
		oldValue = make([]byte, len(value))
		h[string(key)] = oldValue
	} else if len(oldValue) > len(value) {
		h[internal.B2S(key)] = oldValue[:len(value)]
		oldValue = oldValue[:len(value)]
	}

	copy(oldValue, value)
}
