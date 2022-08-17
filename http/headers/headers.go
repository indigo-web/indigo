package headers

import (
	"indigo/errors"
	"indigo/internal"
	"indigo/settings"
)

type ValueAppender func(...byte) int

type HeaderValue struct {
	value []byte
}

func (h HeaderValue) String() string {
	return internal.B2S(h.value)
}

func (h HeaderValue) Bytes() []byte {
	return h.value
}

func (h *HeaderValue) append(chars ...byte) (newLen int) {
	h.value = append(h.value, chars...)
	return len(h.value)
}

type Manager struct {
	headers    map[string]*HeaderValue
	maxHeaders uint8
}

func NewManager(headersSettings settings.HeadersNumber) Manager {
	return Manager{
		headers:    make(map[string]*HeaderValue, headersSettings.Default),
		maxHeaders: headersSettings.Maximal,
	}
}

func (m Manager) Get(key string) (header *HeaderValue, found bool) {
	header, found = m.headers[key]
	return header, found
}

func (m Manager) Set(key []byte) (appender ValueAppender, err error) {
	// TODO: pre-alloc HeaderValue.value slice to some minimal size
	header, found := m.headers[internal.B2S(key)]

	if !found {
		if uint8(len(m.headers)) >= m.maxHeaders {
			return appender, errors.ErrTooManyHeaders
		}

		header = new(HeaderValue)
		m.headers[string(key)] = header
	} else {
		header.value = header.value[:0]
	}

	return header.append, nil
}
