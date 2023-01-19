package http

import "io"

// bodyIOReader is an implementation of io.Reader interface for request body
type bodyIOReader struct {
	reader BodyReader
}

func newBodyIOReader(reader BodyReader) io.Reader {
	return bodyIOReader{
		reader: reader,
	}
}

func (b bodyIOReader) Read(buff []byte) (n int, err error) {
	data, err := b.reader.Read()
	copy(buff, data)

	return len(data), err
}
