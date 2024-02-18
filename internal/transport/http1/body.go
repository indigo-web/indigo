package http1

import (
	"github.com/indigo-web/chunkedbody"
	"github.com/indigo-web/indigo/config"
	"github.com/indigo-web/indigo/http"
	"github.com/indigo-web/indigo/http/status"
	"github.com/indigo-web/indigo/internal/server/tcp"
	"github.com/indigo-web/indigo/internal/stash"
	"github.com/indigo-web/utils/uf"
	"io"
	"math"
)

var _ http.Body = &Body{}

type Body struct {
	*stash.Reader
	plain         plainBodyReader
	chunked       chunkedBodyReader
	isChunked     bool
	contentLength int
	fullBodyBuff  []byte
	eof           bool
}

func NewBody(
	client tcp.Client, chunkedParser *chunkedbody.Parser, s config.Body,
) *Body {
	body := &Body{
		plain:   newPlainBodyReader(client, s.MaxSize),
		chunked: newChunkedBodyReader(client, s.MaxSize, chunkedParser),
	}
	body.Reader = stash.New(body.Retrieve)

	return body
}

func (b *Body) Init(request *http.Request) {
	b.isChunked = request.Encoding.Chunked
	b.contentLength = request.ContentLength
	if b.isChunked {
		b.chunked.init(request)
	} else {
		b.plain.init(request)
	}

	b.eof = false
	b.Reader.Reset()
}

func (b *Body) String() (string, error) {
	bytes, err := b.Bytes()

	return uf.B2S(bytes), err
}

func (b *Body) Bytes() ([]byte, error) {
	if b.eof {
		return b.fullBodyBuff, nil
	}

	// FIXME: this barely works with applied transfer-encodings
	if !b.isChunked && cap(b.fullBodyBuff) < b.contentLength {
		b.fullBodyBuff = make([]byte, 0, b.contentLength)
	}

	b.fullBodyBuff = b.fullBodyBuff[:0]

	for {
		data, err := b.Retrieve()
		b.fullBodyBuff = append(b.fullBodyBuff, data...)
		switch err {
		case nil:
		case io.EOF:
			return b.fullBodyBuff, nil
		default:
			return nil, err
		}
	}
}

func (b *Body) Callback(cb http.OnBodyCallback) error {
	for {
		data, err := b.Retrieve()
		switch err {
		case nil:
		case io.EOF:
			return cb(data)
		default:
			return err
		}

		if err = cb(data); err != nil {
			return err
		}
	}
}

func (b *Body) Retrieve() ([]byte, error) {
	if b.eof {
		return nil, io.EOF
	}

	var (
		piece []byte
		err   error
	)

	if b.isChunked {
		piece, err = b.chunked.read()
	} else {
		piece, err = b.plain.read()
	}

	b.checkEOF(err)

	return piece, err
}

func (b *Body) Discard() (err error) {
	for !b.eof {
		_, err = b.Retrieve()
		if err != nil {
			break
		}
	}

	if err == io.EOF {
		err = nil
	}

	return err
}

func (b *Body) checkEOF(err error) {
	if err == io.EOF {
		b.eof = true
	}
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
	data, err := c.client.Read()
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
	c.client.Unread(extra)

	return chunk, err
}

func adduint(x, y uint) (uint, bool) {
	return x + y, math.MaxUint-x < y
}
