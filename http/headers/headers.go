package headers

// Headers is a struct that encapsulates headers map from user, allowing only
// methods
type Headers struct {
	headers map[string][]string
}

func NewHeaders(underlying map[string][]string) Headers {
	if underlying == nil {
		// underlying MUST NEVER be nil, otherwise this causes panics in different places
		// that are difficult to debug, mostly in tests
		underlying = make(map[string][]string)
	}

	return Headers{
		headers: underlying,
	}
}

// Value does the same as ValueOr does but returning an empty string by default
func (h Headers) Value(key string) string {
	return h.ValueOr(key, "")
}

// ValueOr returns a header value
func (h Headers) ValueOr(key, or string) string {
	if values := h.headers[key]; values != nil {
		return values[0]
	}

	return or
}

// Values returns a slice of values including parameters
func (h Headers) Values(key string) []string {
	return h.headers[key]
}

// Unwrap returns an underlying map as it is. This means that modifying it
// will also affect Headers object
func (h Headers) Unwrap() map[string][]string {
	return h.headers
}

// Add values to the key. In case did not exist, it'll be created
func (h Headers) Add(key string, newValues ...string) {
	h.headers[key] = append(h.headers[key], newValues...)
}

// Set just sets the value of the header to the provided values slice
func (h Headers) Set(key string, values []string) {
	h.headers[key] = values
}

// Has returns true or false depending on whether such a key exists
func (h Headers) Has(key string) bool {
	_, found := h.headers[key]
	return found
}

// Clear headers map. Is a system method, that is not supposed to be ever called by user
func (h Headers) Clear() {
	for k := range h.headers {
		delete(h.headers, k)
	}
}
