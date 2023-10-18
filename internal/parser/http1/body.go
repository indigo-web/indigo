package http1

import (
	"github.com/indigo-web/chunkedbody"
	"github.com/indigo-web/indigo/http"
	"github.com/indigo-web/indigo/http/coding"
	"github.com/indigo-web/indigo/http/headers"
	"github.com/indigo-web/indigo/internal/server/tcp"
	"io"
)

type bodyReader struct {
	client        tcp.Client
	bodyBytesLeft int
	chunkedParser *chunkedbody.Parser
	encoding      headers.Encoding
	manager       coding.Manager
}

func NewBodyReader(
	client tcp.Client, chunkedParser *chunkedbody.Parser, manager coding.Manager,
) http.BodyReader {
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

	if b.encoding.Chunked {
		return b.chunkedBodyReader(data)
	}

	return b.plainBodyReader(data)
}

func (b *bodyReader) plainBodyReader(data []byte) (body []byte, err error) {
	// TODO: we can avoid one extra indirect function call by returning io.EOF with
	//  body at the same time

	b.bodyBytesLeft -= len(data)
	if b.bodyBytesLeft < 0 {
		b.client.Unread(data[len(data)+b.bodyBytesLeft:])
		data = data[:len(data)+b.bodyBytesLeft]
		b.bodyBytesLeft = 0
	}

	for _, token := range b.encoding.Transfer {
		data, err = b.manager.Decode(token, data)
		switch err {
		case nil, io.EOF:
		default:
			return nil, err
		}
	}

	for _, token := range b.encoding.Content {
		data, err = b.manager.Decode(token, data)
		switch err {
		case nil, io.EOF:
		default:
			return nil, err
		}
	}

	return data, nil
}

func (b *bodyReader) chunkedBodyReader(data []byte) (body []byte, err error) {
	for _, token := range b.encoding.Transfer {
		data, err = b.manager.Decode(token, data)
		if err != nil {
			return nil, err
		}
	}

	chunk, extra, err := b.chunkedParser.Parse(data, b.encoding.HasTrailer)
	switch err {
	case nil:
	case io.EOF:
		b.bodyBytesLeft = 0
	default:
		return nil, err
	}

	b.client.Unread(extra)

	for _, token := range b.encoding.Content {
		chunk, err = b.manager.Decode(token, chunk)
		if err != nil {
			return nil, err
		}
	}

	return chunk, nil
}
