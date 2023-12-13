package types

import "io"

// Attachment is a wrapper for io.Reader, with the difference that there is the size attribute.
// If positive value (including 0) is set, then ordinary plain-text response will be rendered.
// Otherwise, chunked transfer encoding is used.
type Attachment struct {
	content io.Reader
	size    int
}

// NewAttachment returns a new Attachment instance
func NewAttachment(content io.Reader, size int) Attachment {
	return Attachment{
		content: content,
		size:    size,
	}
}

func (a Attachment) Content() io.Reader {
	return a.content
}

func (a Attachment) Size() int {
	return a.size
}

func (a Attachment) Close() {
	if closer, ok := a.content.(io.Closer); ok {
		_ = closer.Close()
	}
}
