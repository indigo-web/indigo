package http1

import (
	"io"
	"math"

	"github.com/indigo-web/indigo/config"
	"github.com/indigo-web/indigo/http"
	"github.com/indigo-web/indigo/http/status"
	"github.com/indigo-web/indigo/transport"
)

type body struct {
	maxLen        uint64
	counter       uint64
	reader        func(*body) ([]byte, error)
	chunkedParser chunkedParser
	client        transport.Client
}

func newBody(client transport.Client, s config.Body) *body {
	return &body{
		reader:        nop,
		client:        client,
		maxLen:        s.MaxSize,
		chunkedParser: newChunkedParser(),
	}
}

func (b *body) Fetch() ([]byte, error) {
	return b.reader(b)
}

func (b *body) Reset(request *http.Request) {
	if request.Chunked {
		b.initChunked()
		b.reader = (*body).readChunked
	} else if request.Connection == "close" {
		b.initEOFReader()
		b.reader = (*body).readTillEOF
	} else {
		b.initPlain(uint64(request.ContentLength))
		b.reader = (*body).readPlain
	}
}

func (b *body) initPlain(totalLen uint64) {
	b.counter = totalLen
}

func (b *body) readPlain() (body []byte, err error) {
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

func (b *body) initEOFReader() {
	b.counter = 0
}

func (b *body) readTillEOF() ([]byte, error) {
	chunk, err := b.client.Read()
	if b.counter > math.MaxUint64-uint64(len(chunk)) {
		return nil, status.ErrBodyTooLarge
	}

	b.counter += uint64(len(chunk))

	return chunk, err
}

func (b *body) initChunked() {
	b.counter = 0
}

func (b *body) readChunked() (body []byte, err error) {
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

func nop(*body) ([]byte, error) {
	return nil, io.EOF
}
