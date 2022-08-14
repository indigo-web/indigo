package types

import "bytes"

// ResponseHeaders are built like a slice of byte slices. Key and value
// are just followed one-by-one without any splitters
type ResponseHeaders [][]byte

// Append just appends key and value to headers list
func (r ResponseHeaders) Append(key, value []byte) ResponseHeaders {
	return append(append(r, key), value)
}

func (r ResponseHeaders) Get(key []byte) ([]byte, bool) {
	for i := 0; i < len(r); i += 2 {
		if bytes.Equal(r[i], key) {
			return r[i+1], true
		}
	}

	return nil, false
}

func (r ResponseHeaders) Set(key, value []byte) ResponseHeaders {
	for i := 0; i < len(r); i += 2 {
		if bytes.Equal(r[i], key) {
			r[i+1] = value
			break
		}
	}

	return r
}
