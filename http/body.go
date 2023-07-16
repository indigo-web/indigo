package http

import (
	"github.com/indigo-web/indigo/http/decode"
	"github.com/indigo-web/indigo/http/headers"
	"io"
)

type (
	onBodyCallback func([]byte) error
	BodyReader     interface {
		Init(*Request)
		Read() ([]byte, error)
	}
)

var _ io.Reader = &Body{}

type Body struct {
	reader        BodyReader
	decoder       *decode.Decoder
	te            headers.TransferEncoding
	contentLength int
	bodyBuff      []byte
	rawIOReader   bodyIOReader
}

func NewBody(reader BodyReader, decoder *decode.Decoder) *Body {
	return &Body{
		reader:      reader,
		decoder:     decoder,
		rawIOReader: newBodyIOReader(),
	}
}

func (b *Body) Init(req *Request) {
	b.reader.Init(req)
	b.te = req.TransferEncoding
	b.contentLength = req.ContentLength

	if req.TransferEncoding.Chunked {
		b.contentLength = -1
	}
}

// Raw returns RAW representation of the body. This means, that it may be encoded. If you
// want a decoded body, then use Value() method instead
func (b *Body) Raw() ([]byte, error) {
	if b.bodyBuff == nil && b.contentLength > 0 {
		b.bodyBuff = make([]byte, 0, b.contentLength)
	}

	b.bodyBuff = b.bodyBuff[:0]

	return b.bodyBuff, b.callback(func(data []byte) error {
		b.bodyBuff = append(b.bodyBuff, data...)
		return nil
	})
}

// Value returns decoded full body of the request
func (b *Body) Value() ([]byte, error) {
	rawData, err := b.Raw()
	if err != nil {
		return nil, err
	}

	return b.decoder.Decode(b.te.Token, rawData)
}

// Read implements the io.Reader interface, so behaves respectively
func (b *Body) Read(buff []byte) (n int, err error) {
	rawData, err := b.reader.Read()
	if err != nil {
		return 0, err
	}

	data, err := b.decoder.Decode(b.te.Token, rawData)
	copy(buff, data)

	return len(data), err
}

// RawReader returns io.Reader implementation in order to read directly from the input
// stream. The stream may be encoded, so highly recommended just to use the Body entity
// as io.Reader implementation
func (b *Body) RawReader() io.Reader {
	b.rawIOReader.Reassign(b.reader)

	return b.rawIOReader
}

// Callback takes a function, that'll be called with body piece every time it's received.
// In case error is returned from the callback, it'll also be returned from this method
func (b *Body) Callback(onBody onBodyCallback) error {
	return b.callback(func(data []byte) error {
		decoded, err := b.decoder.Decode(b.te.Token, data)
		if err != nil {
			return err
		}

		return onBody(decoded)
	})
}

// RawCallback does the same as Callback, but body pieces aren't prematurely decoded.
func (b *Body) RawCallback(onBody onBodyCallback) error {
	return b.callback(onBody)
}

func (b *Body) callback(onBody onBodyCallback) error {
	for {
		piece, err := b.reader.Read()
		switch err {
		case nil:
		case io.EOF:
			return nil
		default:
			return err
		}
		if err = onBody(piece); err != nil {
			return err
		}
	}
}

func (b *Body) Reset() error {
	for {
		_, err := b.reader.Read()
		switch err {
		case nil:
		case io.EOF:
			return nil
		default:
			return err
		}
	}
}
