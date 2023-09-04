package headers

type Encoding struct {
	// Transfer represents Transfer-Encoding header value
	Transfer struct {
		Tokens []string
	}

	// Content represents Content-Encoding header value
	Content struct {
		Tokens []string
	}

	// Chunked doesn't belong to any of encodings, as it is still must be processed individually
	Chunked, HasTrailer bool
}
