package coding

type Encoding struct {
	// Transfer stores Transfer-Encoding tokens
	Transfer []string
	// Content stores Content-Encoding tokens
	Content []string
	// HasTrailer defines, whether there are additional headers can be stored in
	// a zero-length chunk (if Transfer-Encoding is chunked)
	HasTrailer bool
}

// Identity returns, whether the message was encoded
func (e Encoding) Identity() bool {
	return len(e.Transfer) == 0 && len(e.Content) == 0
}
