package webserver

import (
	"bytes"
	"github.com/fakefloordiv/snowdrop-http/httpparser"
)

type Header struct {
	Key   []byte
	Value []byte
}

type Headers struct {
	headers []Header
}

/*
AppendAssertDuplicate appends a header to the list of the headers. Returns bool that means whether
there is already a header with such a key.

Not recommended to use manually, as this function is for internal stuff

Also, this function affects the Key field of the header structure is passed by lower-casing
all the characters in it
*/
func (h *Headers) AppendAssertDuplicate(header Header) (isDuplicate bool) {
	if isHeaderDuplicate(header.Key, h.headers) {
		return true
	}

	toLowercase(header.Key)
	h.headers = append(h.headers, header)

	return false
}

/*
Get by name that is a usual bytes array. In case there is no such a header, nil will be returned
*/
func (h Headers) Get(name []byte) []byte {
	for _, header := range h.headers {
		if bytes.Equal(header.Key, name) {
			return header.Value
		}
	}

	return nil
}

/*
GetString is just a wrapper on Get, but uses string as a type that may be a bit more
convenient in some cases

Returns 2 values as string is not an array, it can't be nil, so error must be returned
*/
func (h Headers) GetString(name string) (string, error) {
	header := h.Get([]byte(name))

	if header == nil {
		return "", ErrHeaderNotFound
	}

	return string(header), nil
}

/*
GetStringUnsafe is the same as usual GetString, but uses unsafe s2b & b2s functions instead

s2b & b2s functions are defined in unsafefeatures.go file
*/
func (h Headers) GetStringUnsafe(name string) (string, error) {
	header := h.Get(s2b(name))

	if header == nil {
		return "", ErrHeaderNotFound
	}

	return b2s(header), nil
}

func isHeaderDuplicate(key []byte, headers []Header) bool {
	for _, header := range headers {
		if httpparser.EqualFold(header.Key, key) {
			return true
		}
	}

	return false
}
