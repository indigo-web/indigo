package headers

type Iterator[T any] func() (element T, continue_ bool)

// Headers is a struct that encapsulates headers map from user, allowing only
// methods
type Headers struct {
	headers []string
	iterTmp []string
}

func NewHeaders(underlying map[string][]string) *Headers {
	if underlying == nil {
		// underlying MUST NEVER be nil, otherwise this causes panics in different places
		// that are difficult to debug, mostly in tests
		underlying = make(map[string][]string)
	}

	return &Headers{
		headers: map2slice(underlying),
	}
}

func (h *Headers) KeysIter() Iterator[string] {
	// TODO: implement the same method, but using arenas
	var index int

	return func() (element string, continue_ bool) {
		if index >= len(h.headers) {
			return element, false
		}

		element = h.headers[index]
		index += 2

		return element, true
	}
}

// Value does the same as ValueOr does but returning an empty string by default
func (h *Headers) Value(key string) string {
	return h.ValueOr(key, "")
}

// ValueOr returns a header value
func (h *Headers) ValueOr(key, or string) string {
	for i := 0; i < len(h.headers); i += 2 {
		if h.headers[i] == key {
			return h.headers[i+1]
		}
	}

	return or
}

func (h *Headers) ValuesIter(key string) Iterator[string] {
	var offset int

	return func() (element string, continue_ bool) {
		if offset >= len(h.headers) {
			return "", false
		}

		for ; offset < len(h.headers); offset += 2 {
			if h.headers[offset] == key {
				return h.headers[offset+1], true
			}
		}

		return "", false
	}
}

// Values returns a slice of values including parameters
func (h *Headers) Values(key string) (values []string) {
	// TODO: amortize allocations by using arena
	for i := 0; i < len(h.headers); i += 2 {
		if h.headers[i] == key {
			values = append(values, h.headers[i+1])
		}
	}

	return values
}

// Unwrap returns an underlying map as it is. This means that modifying it
// will also affect Headers object
func (h *Headers) Unwrap() []string {
	return h.headers
}

// Add values to the key. In case did not exist, it'll be created
func (h *Headers) Add(key string, newValues ...string) {
	for i := range newValues {
		h.headers = append(h.headers, key, newValues[i])
	}
}

// Has returns true or false depending on whether such a key exists
func (h *Headers) Has(key string) bool {
	for i := 0; i < len(h.headers); i += 2 {
		if h.headers[i] == key {
			return true
		}
	}

	return false
}

// Clear headers map. Is a system method, that is not supposed to be ever called by user
func (h *Headers) Clear() {
	h.headers = h.headers[:0]
}

func map2slice(m map[string][]string) []string {
	headers := make([]string, 0, len(m))

	for key, values := range m {
		for _, value := range values {
			headers = append(headers, key, value)
		}
	}

	return headers
}
