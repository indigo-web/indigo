package http1

import (
	"github.com/indigo-web/chunkedbody"
	"github.com/indigo-web/indigo/http"
	"github.com/indigo-web/indigo/http/coding"
	"github.com/indigo-web/indigo/http/headers"
	"github.com/indigo-web/indigo/internal/server/tcp"
	"github.com/indigo-web/utils/uf"
	"github.com/indigo-web/utils/unreader"
	"io"
)

var _ http.Body = &Body{}

const chunkedMode = -1

type Body struct {
	client        tcp.Client
	bodyBytesLeft int
	chunkedParser *chunkedbody.Parser
	encoding      headers.Encoding
	manager       coding.Manager
	unreader      unreader.Unreader
	fullBodyBuff  []byte
}

func NewBody(
	client tcp.Client, chunkedParser *chunkedbody.Parser, manager coding.Manager,
) *Body {
	return &Body{
		client:        client,
		chunkedParser: chunkedParser,
		manager:       manager,
	}
}

func (b *Body) Init(request *http.Request) {
	b.encoding = request.Encoding
	b.bodyBytesLeft = request.ContentLength
	if request.Encoding.Chunked {
		b.bodyBytesLeft = chunkedMode
	}
}

func (b *Body) Read(into []byte) (n int, err error) {
	data, err := b.Retrieve()
	n = copy(into, data)
	b.unreader.Unread(data[n:])

	return n, err
}

func (b *Body) String() (string, error) {
	bytes, err := b.Bytes()

	return uf.B2S(bytes), err
}

func (b *Body) Bytes() ([]byte, error) {
	if b.bodyBytesLeft != chunkedMode && cap(b.fullBodyBuff) < b.bodyBytesLeft {
		b.fullBodyBuff = make([]byte, 0, b.bodyBytesLeft)
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

func (b *Body) Reset() error {
	for b.bodyBytesLeft > 0 {
		_, err := b.Retrieve()
		// this structural piece of code exists here just because of irrational fear
		// of unconditional jumps. I could make this prettier by pasting ordinary
		// error-switch as I used to do in all the samples above (and probably below), but
		// from the switch I wouldn't be able to break a loop. So instead, we've got this
		// pasta italiana. Enjoy the world you created.
		if err != nil {
			if err == io.EOF {
				break
			}

			return err
		}
	}

	b.unreader.Reset()

	return nil
}

func (b *Body) Retrieve() ([]byte, error) {
	return b.unreader.PendingOr(func() ([]byte, error) {
		if b.bodyBytesLeft == 0 {
			return nil, io.EOF
		}

		data, err := b.client.Read()
		if err != nil {
			return nil, err
		}

		if b.encoding.Chunked {
			return b.chunkedBody(data)
		}

		return b.plainBody(data)
	})
}

func (b *Body) plainBody(data []byte) (body []byte, err error) {
	// TODO: we can avoid one extra indirect function call by returning io.EOF with
	//  body at the same time

	b.bodyBytesLeft -= len(data)
	if b.bodyBytesLeft < 0 {
		b.client.Unread(data[len(data)+b.bodyBytesLeft:])
		data = data[:len(data)+b.bodyBytesLeft]
		b.bodyBytesLeft = 0
	}

	return data, nil
}

func (b *Body) chunkedBody(data []byte) (body []byte, err error) {
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

	return chunk, nil
}
