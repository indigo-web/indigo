package http1

import (
	"github.com/indigo-web/indigo/config"
	"github.com/indigo-web/indigo/http"
	"github.com/indigo-web/indigo/http/status"
	"github.com/indigo-web/indigo/transport"
	"io"
	"math"
)

type Body struct {
	maxLen        uint64
	counter       uint64
	reader        func() ([]byte, error)
	chunkedParser chunkedParser
	client        transport.Client
}

func NewBody(client transport.Client, s config.Body) *Body {
	return &Body{
		reader:        nop,
		client:        client,
		maxLen:        s.MaxSize,
		chunkedParser: newChunkedParser(),
	}
}

func (b *Body) Fetch() ([]byte, error) {
	return b.reader()
}

func (b *Body) Reset(request *http.Request) {
	if request.Encoding.Chunked {
		b.initChunked()
		b.reader = b.readChunked
	} else if request.Connection == "close" {
		b.initEOFReader()
		b.reader = b.readTillEOF
	} else {
		b.initPlain(uint64(request.ContentLength))
		b.reader = b.readPlain
	}
}

func (b *Body) initPlain(totalLen uint64) {
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

	if uint64(len(data)) >= b.counter {
		body, data = data[:b.counter], data[b.counter:]
		b.client.Pushback(data)
		b.counter = 0
		err = io.EOF
	} else {
		b.counter -= uint64(len(data))
		body = data
	}

	return body, err
}

func (b *Body) initEOFReader() {
	b.counter = 0
}

func (b *Body) readTillEOF() ([]byte, error) {
	chunk, err := b.client.Read()
	if b.counter > math.MaxUint64-uint64(len(chunk)) {
		return nil, status.ErrBodyTooLarge
	}

	b.counter += uint64(len(chunk))

	return chunk, err
}

func (b *Body) initChunked() {
	b.counter = 0
}

func (b *Body) readChunked() (body []byte, err error) {
	data, err := b.client.Read()
	if err != nil {
		return nil, err
	}

	chunk, extra, err := b.chunkedParser.Parse(data)
	switch err {
	case nil, io.EOF:
	default:
		return nil, err
	}

	if b.counter > math.MaxUint64-uint64(len(chunk)) {
		return nil, status.ErrBodyTooLarge
	}

	b.counter += uint64(len(chunk))
	b.client.Pushback(extra)

	return chunk, err
}

func nop() ([]byte, error) {
	return nil, io.EOF
}
