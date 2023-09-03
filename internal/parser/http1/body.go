package http1

import (
	"github.com/indigo-web/indigo/http"
	"github.com/indigo-web/indigo/http/decoder"
	"github.com/indigo-web/indigo/http/headers"
	"github.com/indigo-web/indigo/internal/server/tcp"
	"io"
)

type bodyReader struct {
	client        tcp.Client
	bodyBytesLeft int
	chunkedParser *ChunkedBodyParser
	encoding      headers.Encoding
	manager       *decoder.Manager
}

func NewBodyReader(client tcp.Client, chunkedParser *ChunkedBodyParser, manager *decoder.Manager) http.BodyReader {
	return &bodyReader{
		client:        client,
		chunkedParser: chunkedParser,
		manager:       manager,
	}
}

const chunkedMode = -1

func (b *bodyReader) Init(request *http.Request) {
	b.encoding = request.Encoding

	if !request.Encoding.Chunked {
		b.bodyBytesLeft = request.ContentLength
		return
	}

	b.bodyBytesLeft = chunkedMode
}

func (b *bodyReader) Read() ([]byte, error) {
	if b.bodyBytesLeft == 0 {
		return nil, io.EOF
	}

	data, err := b.client.Read()
	if err != nil {
		return nil, err
	}

	if len(b.encoding.Tokens) != 0 {
		data = b.plainBodyReader(data)

		for _, token := range b.encoding.Tokens {
			data, err = b.manager.Decode(token, data)
			if err != nil {
				return nil, err
			}
		}

		if !b.encoding.Chunked {
			return data, nil
		}
	}

	if b.encoding.Chunked {
		return b.chunkedBodyReader(data)
	}

	return b.plainBodyReader(data), nil
}

func (b *bodyReader) plainBodyReader(data []byte) []byte {
	b.bodyBytesLeft -= len(data)
	if b.bodyBytesLeft < 0 {
		b.client.Unread(data[len(data)+b.bodyBytesLeft:])
		data = data[:len(data)+b.bodyBytesLeft]
		b.bodyBytesLeft = 0
	}

	return data
}

func (b *bodyReader) chunkedBodyReader(data []byte) ([]byte, error) {
	chunk, extra, err := b.chunkedParser.Parse(data, b.encoding.HasTrailer)
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
