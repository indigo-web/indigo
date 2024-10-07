package http1

import (
	"github.com/indigo-web/chunkedbody"
	"github.com/indigo-web/indigo/config"
	"github.com/indigo-web/indigo/http"
	"github.com/indigo-web/indigo/http/status"
	"github.com/indigo-web/indigo/internal/tcp"
	"io"
	"math"
)

type Body struct {
	plain         plainBodyReader
	chunked       chunkedBodyReader
	isChunked     bool
	contentLength int
}

func NewBody(
	client tcp.Client, chunkedParser *chunkedbody.Parser, s config.Body,
) *Body {
	return &Body{
		plain:   newPlainBodyReader(client, s.MaxSize),
		chunked: newChunkedBodyReader(client, s.MaxSize, chunkedParser),
	}
}

func (b *Body) Init(request *http.Request) {
	b.isChunked = request.Encoding.Chunked
	b.contentLength = request.ContentLength
	if b.isChunked {
		b.chunked.init(request)
	} else {
		b.plain.init(request)
	}
}

func (b *Body) Retrieve() ([]byte, error) {
	var (
		piece []byte
		err   error
	)

	if b.isChunked {
		piece, err = b.chunked.read()
	} else {
		piece, err = b.plain.read()
	}

	return piece, err
}

type plainBodyReader struct {
	client                tcp.Client
	maxBodyLen, bytesLeft uint
}

func newPlainBodyReader(client tcp.Client, maxBodyLen uint) plainBodyReader {
	return plainBodyReader{
		client:     client,
		maxBodyLen: maxBodyLen,
	}
}

func (p *plainBodyReader) init(request *http.Request) {
	p.bytesLeft = uint(request.ContentLength)
}

func (p *plainBodyReader) read() (body []byte, err error) {
	if p.bytesLeft == 0 {
		return nil, io.EOF
	}

	data, err := p.client.Read()
	if err != nil {
		return nil, err
	}

	if p.bytesLeft > p.maxBodyLen {
		return nil, status.ErrBodyTooLarge
	}

	if dataLen := uint(len(data)); dataLen >= p.bytesLeft {
		body, data = data[:p.bytesLeft], data[p.bytesLeft:]
		p.client.Unread(data)
		p.bytesLeft = 0
		err = io.EOF
	} else {
		p.bytesLeft -= dataLen
		body = data
	}

	return body, err
}

type chunkedBodyReader struct {
	client               tcp.Client
	maxBodyLen, received uint
	encoding             http.Encoding
	parser               *chunkedbody.Parser
}

func newChunkedBodyReader(client tcp.Client, maxBodyLen uint, parser *chunkedbody.Parser) chunkedBodyReader {
	return chunkedBodyReader{
		client:     client,
		maxBodyLen: maxBodyLen,
		parser:     parser,
	}
}

func (c *chunkedBodyReader) init(request *http.Request) {
	c.encoding = request.Encoding
	c.received = 0
}

func (c *chunkedBodyReader) read() (body []byte, err error) {
	client := c.client
	data, err := client.Read()
	if err != nil {
		return nil, err
	}

	chunk, extra, err := c.parser.Parse(data, c.encoding.HasTrailer)
	switch err {
	case nil, io.EOF:
	default:
		return nil, err
	}

	received, overflows := adduint(c.received, uint(len(chunk)))
	if overflows || received > c.maxBodyLen {
		return nil, status.ErrBodyTooLarge
	}

	c.received = received
	client.Unread(extra)

	return chunk, err
}

func adduint(x, y uint) (uint, bool) {
	return x + y, math.MaxUint-x < y
}
