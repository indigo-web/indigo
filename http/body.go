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

func (b *Body) Value() ([]byte, error) {
	encoded, err := b.Raw()
	if err != nil {
		return nil, err
	}

	return b.decoder.Decode(b.te.Token, encoded)
}

func (b *Body) Read(buff []byte) (n int, err error) {
	data, err := b.reader.Read()
	copy(buff, data)

	return len(data), err
}

func (b *Body) RawReader() io.Reader {
	b.rawIOReader.Reset()

	return b.rawIOReader
}

func (b *Body) Callback(onBody onBodyCallback) error {
	return b.callback(func(data []byte) error {
		decoded, err := b.decoder.Decode(b.te.Token, data)
		if err != nil {
			return err
		}

		return onBody(decoded)
	})
}

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
