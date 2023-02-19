package http1

import (
	"io"

	"github.com/indigo-web/indigo/http"
	"github.com/indigo-web/indigo/internal/server/tcp"
	"github.com/indigo-web/indigo/settings"
)

type bodyReader struct {
	client             tcp.Client
	bodyBytesLeft      int
	chunkedParser      chunkedBodyParser
	chunkedBodyTrailer bool
}

func NewBodyReader(client tcp.Client, bodySettings settings.Body) http.BodyReader {
	return &bodyReader{
		client:        client,
		chunkedParser: newChunkedBodyParser(bodySettings),
	}
}

func (b *bodyReader) Init(request *http.Request) {
	if !request.IsChunked {
		b.bodyBytesLeft = request.ContentLength
		return
	}

	b.bodyBytesLeft = -1
	b.chunkedBodyTrailer = request.Headers.Has("trailer")
}

func (b *bodyReader) Read() ([]byte, error) {
	switch b.bodyBytesLeft {
	case 0:
		return nil, io.EOF
	case -1:
		return b.chunkedBodyReader()
	default:
		return b.plainBodyReader()
	}
}

func (b *bodyReader) plainBodyReader() ([]byte, error) {
	data, err := b.client.Read()

	b.bodyBytesLeft -= len(data)
	if b.bodyBytesLeft < 0 {
		b.client.Unread(data[len(data)+b.bodyBytesLeft-1:])
		data = data[:len(data)+b.bodyBytesLeft-1]
		b.bodyBytesLeft = 0
	}

	return data, err
}

func (b *bodyReader) chunkedBodyReader() ([]byte, error) {
	data, err := b.client.Read()
	if err != nil {
		return nil, err
	}

	chunk, extra, err := b.chunkedParser.Parse(data, b.chunkedBodyTrailer)
	switch err {
	case nil:
	case io.EOF:
		b.bodyBytesLeft = 0
	default:
		return nil, err
	}

	b.client.Unread(extra)

	return chunk, nil
}
