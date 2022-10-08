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

func (h Headers) AsMap() map[string][]string {
	return h.headers
}

func (h Headers) Add(key string, newValues ...string) {
	h.headers[key] = append(h.headers[key], newValues...)
}

func (h Headers) Set(key string, values []string) {
	h.headers[key] = values
}

func (h Headers) Clear() {
	for k := range h.headers {
		delete(h.headers, k)
	}
}
