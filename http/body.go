package http

import (
	"io"
	"slices"

	"github.com/flrdv/uf"
	"github.com/indigo-web/indigo/http/form"
	"github.com/indigo-web/indigo/http/mime"
	"github.com/indigo-web/indigo/http/status"
	"github.com/indigo-web/indigo/internal/formdata"
	"github.com/indigo-web/indigo/internal/strutil"
	json "github.com/json-iterator/go"
)

// Fetcher abstracts the underlying protocol-dependant body source. Even though the signature
// is identical to transport.Client.Read(), it is named differently in order to highlight the
// difference between relatively low-level connection "raw data" and high-level body
// data streams.
type Fetcher interface {
	Fetch() ([]byte, error)
}

type Body struct {
	Fetcher

	request  *Request
	error    error
	buff     []byte
	formbuff []byte
	pending  []byte
	form     form.Form
}

func NewBody(src Fetcher) *Body {
	return &Body{
		Fetcher: src,
	}
}

// Callback invokes the callback every time as there's a piece of body available
// for reading. If the callback returns an error, it'll be passed back to the caller.
// The callback is not notified when there's no more data or networking error has
// occurred.
//
// Please note: this method can be used only once.
func (b *Body) Callback(cb func([]byte) error) error {
	if b.error != nil {
		return b.error
	}

	for {
		var data []byte
		data, b.error = b.Fetch()
		switch b.error {
		case nil:
		case io.EOF:
			if len(data) > 0 {
				return cb(data)
			}

			return nil
		default:
			return b.error
		}

		if b.error = cb(data); b.error != nil {
			return b.error
		}
	}
}

// Bytes returns the whole body at once in a byte representation.
func (b *Body) Bytes() ([]byte, error) {
	if len(b.buff) != 0 {
		return b.buff, nil
	}

	if b.error != nil {
		return nil, b.error
	}

	newSize := int(b.request.cfg.Body.Form.BufferPrealloc)
	if !b.request.Chunked {
		newSize = min(b.request.ContentLength, int(b.request.cfg.Body.MaxSize))
	}

	b.buff = slices.Grow(b.buff[:0], newSize)

	for {
		var data []byte
		data, b.error = b.Fetch()
		b.buff = append(b.buff, data...)
		switch b.error {
		case nil:
		case io.EOF:
			return b.buff, nil
		default:
			return nil, b.error
		}
	}
}

// String returns the whole body at once in a string representation.
func (b *Body) String() (string, error) {
	bytes, err := b.Bytes()
	return uf.B2S(bytes), err
}

// Read implements the io.Reader interface.
func (b *Body) Read(into []byte) (n int, err error) {
	if len(b.pending) == 0 && b.error == nil {
		b.pending, b.error = b.Fetch()
	}

	n = copy(into, b.pending)
	b.pending = b.pending[n:]

	if len(b.pending) == 0 && b.error != nil {
		err = b.error
	}

	return n, err
}

// JSON convoys the request's body to a json unmarshaller automatically and behaves
// in a similar manner.
//
// Please note: this method cannot be used on requests with Content-Type incompatible
// with mime.JSON (in this case, status.ErrUnsupportedMediaType is returned).
//
// TODO: make possible to choose and use different from json-iterator json marshall/unmarshall
func (b *Body) JSON(model any) error {
	if !mime.Complies(mime.JSON, b.request.ContentType) {
		return status.ErrUnsupportedMediaType
	}

	data, err := b.Bytes()
	if err != nil {
		return err
	}

	iterator := json.ConfigDefault.BorrowIterator(data)
	iterator.ReadVal(model)
	err = iterator.Error
	json.ConfigDefault.ReturnIterator(iterator)

	return err
}

// Form interprets the request's body as a mime.FormUrlencoded data and
// returns parsed key-value pairs. If the request's MIME type is defined and is different
// from mime.FormUrlencoded, status.ErrUnsupportedMediaType is returned
func (b *Body) Form() (f form.Form, err error) {
	if b.form == nil {
		b.form = make(form.Form, b.request.cfg.Body.Form.EntriesPrealloc)
	}
	if b.formbuff == nil {
		b.formbuff = make([]byte, b.request.cfg.Body.Form.BufferPrealloc)
	}

	raw, err := b.Bytes()
	if err != nil {
		return nil, err
	}

	switch {
	case mime.Complies(mime.FormUrlencoded, b.request.ContentType):
		f, b.formbuff, err = formdata.ParseFormURLEncoded(b.form[:0], raw, b.formbuff[:0])
		return f, err
	case mime.Complies(mime.Multipart, b.request.ContentType):
		boundary, ok := b.multipartBoundary()
		if !ok {
			return nil, status.ErrBadRequest
		}

		return formdata.ParseMultipart(b.request.cfg, b.form[:0], raw, b.formbuff[:0], boundary)
	default:
		return nil, status.ErrUnsupportedMediaType
	}
}

func (b *Body) Len() int {
	if b.request.Chunked {
		return -1
	}

	return b.request.ContentLength
}

// Discard sinkholes the rest of the body. Should not be used unless you know what you're doing.
func (b *Body) Discard() error {
	for b.error == nil {
		_, b.error = b.Fetch()
	}

	if b.error == io.EOF {
		return nil
	}

	return b.error
}

// Reset resets the body state. Should never be used as serves internal purposes only.
func (b *Body) Reset(request *Request) {
	b.error = nil
	b.buff = b.buff[:0]
	b.pending = b.pending[:0]
	b.request = request
}

func (b *Body) multipartBoundary() (boundary string, ok bool) {
	for key, value := range strutil.WalkKV(strutil.CutParams(b.request.ContentType)) {
		if key == "boundary" {
			return value, true
		}
	}

	return "", false
}
