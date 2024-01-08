package http

type Encoding struct {
	// Transfer represents Transfer-Encoding header value, split by comma
	Transfer []string
	// Content represents Content-Encoding header value, split by comma
	Content []string
	// Chunked doesn't belong to any of encodings, as it is still must be processed individually
	Chunked, HasTrailer bool
}

// Identity returns, whether the message was encoded
func (e Encoding) Identity() bool {
	return len(e.Transfer) == 0 && len(e.Content) == 0
}
