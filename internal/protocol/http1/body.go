package http1

import (
	"github.com/indigo-web/chunkedbody"
	"github.com/indigo-web/indigo/config"
	"github.com/indigo-web/indigo/http"
	"github.com/indigo-web/indigo/http/status"
	"github.com/indigo-web/indigo/transport"
	"io"
	"math"
)

type chunkedBodyReader struct {
	parser     *chunkedbody.Parser
	hasTrailer bool
}

type Body struct {
	reader  func() ([]byte, error)
	client  transport.Client
	maxLen  uint
	counter uint
	chunked chunkedBodyReader
}

func NewBody(client transport.Client, chunkedParser *chunkedbody.Parser, s config.Body) *Body {
	return &Body{
		reader:  nop,
		client:  client,
		maxLen:  s.MaxSize,
		chunked: newChunkedBodyReader(chunkedParser),
	}
}

func (b *Body) Retrieve() ([]byte, error) {
	return b.reader()
}

func (b *Body) Reset(request *http.Request) {
	if request.Encoding.Chunked {
		b.initChunked(request.Encoding.HasTrailer)
		b.reader = b.readChunked
	} else if request.Connection == "close" {
		b.initEOFReader()
		b.reader = b.readTillEOF
	} else {
		b.initPlain(uint(request.ContentLength))
		b.reader = b.readPlain
	}
}

func (b *Body) initPlain(totalLen uint) {
	b.counter = totalLen
}

func (b *Body) readPlain() (body []byte, err error) {
	if b.counter == 0 {
		return nil, io.EOF
	}

	if b.counter > b.maxLen {
		return nil, status.ErrBodyTooLarge
	}

	data, err := b.client.Read()
	if err != nil {
		return nil, err
	}

	if uint(len(data)) >= b.counter {
		body, data = data[:b.counter], data[b.counter:]
		b.client.Unread(data)
		b.counter = 0
		err = io.EOF
	} else {
		b.counter -= uint(len(data))
		body = data
	}

	return body, err
}

func (b *Body) initEOFReader() {
	b.counter = 0
}

func (b *Body) readTillEOF() ([]byte, error) {
	chunk, err := b.client.Read()
	if b.counter > math.MaxUint-uint(len(chunk)) {
		return nil, status.ErrBodyTooLarge
	}

	b.counter += uint(len(chunk))

	return chunk, err
}

func newChunkedBodyReader(parser *chunkedbody.Parser) chunkedBodyReader {
	return chunkedBodyReader{
		parser: parser,
	}
}

func (b *Body) initChunked(hasTrailer bool) {
	b.chunked.hasTrailer = hasTrailer
	b.counter = 0
}

func (b *Body) readChunked() (body []byte, err error) {
	data, err := b.client.Read()
	if err != nil {
		return nil, err
	}

	chunk, extra, err := b.chunked.parser.Parse(data, b.chunked.hasTrailer)
	switch err {
	case nil, io.EOF:
	default:
		return nil, err
	}

	if b.counter > math.MaxUint-uint(len(chunk)) {
		return nil, status.ErrBodyTooLarge
	}

	b.counter += uint(len(chunk))
	b.client.Unread(extra)

	return chunk, err
}

func nop() ([]byte, error) {
	return nil, io.EOF
}
