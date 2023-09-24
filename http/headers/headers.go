package headers

import (
	"github.com/indigo-web/iter"
	"github.com/indigo-web/utils/strcomp"
)

// Headers is a struct that encapsulates headers map from user, allowing only
// methods
type Headers struct {
	headers []string
	// each method must use its own buffer, as situation when using both at the same
	// time is not that rare. Although, they must stay nil till first usage, to avoid
	// allocating memory for unused features. Although it costs one branch
	keysIterBuff, valuesIterBuff, uniqueBuff []string
}

// FromMap returns headers instance with already inserted values from given map.
// Note: as maps are unordered, resulting Headers instance will also contain
// unordered pairs (but guaranteed to be grouped together)
func FromMap(m map[string][]string) *Headers {
	// this actually doesn't always allocate exactly enough sized slice, as we don't
	// count amount of _values_, only _keys_, where each key may contain more  (or less)
	// than 1 value. But this doesn't actually matter, as this job is made just once
	// per client, so considered not to be a hot path
	headers := make([]string, 0, len(m))

	for key, values := range m {
		for _, value := range values {
			headers = append(headers, key, value)
		}
	}

	return &Headers{
		headers: headers,
	}
}

// NewPreallocHeaders returns empty Headers instance with pre-allocated storage for
// exactly n pairs (two different values of a single header are also counted as two
// pairs)
func NewPreallocHeaders(n int) *Headers {
	return &Headers{
		headers: make([]string, 0, n*2),
	}
}

// NewHeaders returns new empty instance of Headers
func NewHeaders() *Headers {
	return new(Headers)
}

// Add values to the key. In case did not exist, it'll be created
func (h *Headers) Add(key, value string) {
	h.headers = append(h.headers, key, value)
}

// Get behaves just as map lookup. It returns both desired string and bool flag meaning the success
// of the operation (false=no such key, true=found)
func (h *Headers) Get(key string) (string, bool) {
	value := h.Value(key)
	if len(value) == 0 {
		return "", false
	}

	return value, true
}

// Value does the same as ValueOr does but returning an empty string by default
func (h *Headers) Value(key string) string {
	return h.ValueOr(key, "")
}

// ValueOr returns a header value, or custom value instead
func (h *Headers) ValueOr(key, or string) string {
	for i := 0; i < len(h.headers); i += 2 {
		if strcomp.EqualFold(h.headers[i], key) {
			return h.headers[i+1]
		}
	}

	return or
}

// ValuesIter returns an iterator over all the values of a key
func (h *Headers) ValuesIter(key string) iter.Iterator[string] {
	valuesPairs := iter.Filter[[]string](h.Iter(), func(el []string) bool {
		return strcomp.EqualFold(el[0], key)
	})

	return iter.Map[[]string, string](valuesPairs, func(el []string) string {
		return el[1]
	})
}

// Values returns all the values of a key
func (h *Headers) Values(key string) (values []string) {
	values = iter.Extract(h.ValuesIter(key), h.ensureNotNil(h.valuesIterBuff))
	h.valuesIterBuff = values[:0]

	return values
}

// KeysIter returns an iterator over all unique keys
func (h *Headers) KeysIter() iter.Iterator[string] {
	keys := iter.Map[[]string, string](h.Iter(), func(el []string) string {
		return el[0]
	})
	buff := h.ensureNotNil(h.uniqueBuff)
	it := iter.Filter[string](keys, func(el string) bool {
		if contains(buff, el) {
			return false
		}

		buff = append(buff, el)
		return true
	})

	h.uniqueBuff = buff[:0]

	return it
}

// Keys returns all the unique keys
func (h *Headers) Keys() []string {
	keys := iter.Extract(h.KeysIter(), h.ensureNotNil(h.keysIterBuff))
	h.keysIterBuff = keys[:0]

	return keys
}

// Iter returns an iterator over all the header key-value pairs (each pair is
// exactly 2 values). Note: these pairs aren't sorted nor unique, so multiple
// Constant-Header: <some value> pairs may have any other pairs between
func (h *Headers) Iter() iter.Iterator[[]string] {
	return iter.PairedSlice(h.headers)
}

// Has returns true or false depending on whether such a key exists
func (h *Headers) Has(key string) bool {
	for i := 0; i < len(h.headers); i += 2 {
		if strcomp.EqualFold(h.headers[i], key) {
			return true
		}
	}

	return false
}

// Unwrap returns an underlying map as it is. This means that modifying it
// will also affect Headers object
func (h *Headers) Unwrap() []string {
	return h.headers
}

// Clear headers map. Is a system method, that is not supposed to be ever called by user
func (h *Headers) Clear() {
	h.headers = h.headers[:0]
}

func (h *Headers) ensureNotNil(buff []string) []string {
	if buff == nil {
		buff = make([]string, 0, len(h.headers)/2)
	}

	return buff
}

func contains(elements []string, key string) bool {
	for i := range elements {
		if strcomp.EqualFold(elements[i], key) {
			return true
		}
	}

	return false
}
