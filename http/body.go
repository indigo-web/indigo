package http

import (
	"github.com/indigo-web/indigo/config"
	"github.com/indigo-web/indigo/http/form"
	"github.com/indigo-web/indigo/http/mime"
	"github.com/indigo-web/indigo/http/status"
	"github.com/indigo-web/indigo/internal/formdata"
	"github.com/indigo-web/indigo/internal/strutil"
	"github.com/indigo-web/utils/uf"
	json "github.com/json-iterator/go"
	"io"
)

type BodyCallback func([]byte) error

type Retriever interface {
	// Retrieve reads and returns a piece of body available for processing
	Retrieve() ([]byte, error)
}

type retriever = Retriever

type Body struct {
	retriever
	request  *Request
	cfg      *config.Config
	form     form.Form
	buff     []byte
	formbuff []byte
	pending  []byte
	error    error
}

func NewBody(r *Request, impl retriever, cfg *config.Config) *Body {
	return &Body{
		retriever: impl,
		request:   r,
		cfg:       cfg,
	}
}

// Callback invokes the callback every time as there's a piece of body available
// for reading. If the callback returns an error, it'll be passed back to the caller.
// The callback is not notified when there's no more data or networking error has
// occurred.
//
// Please note: this method can be used only once.
func (b *Body) Callback(cb BodyCallback) error {
	if b.error != nil {
		return b.error
	}

	for {
		var data []byte
		data, b.error = b.Retrieve()
		switch b.error {
		case nil:
		case io.EOF:
			return cb(data)
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

	if b.buff == nil {
		b.buff = make([]byte, 0, b.cfg.Body.Form.BufferPrealloc)
	}

	for {
		var data []byte
		data, b.error = b.Retrieve()
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
		b.pending, b.error = b.Retrieve()
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
	// lazily allocate both in order to avoid wasting RAM when we don't really need it.
	// Must add no runtime penalty, as the operation is done once
	if b.form == nil {
		b.form = make(form.Form, b.cfg.Body.Form.EntriesPrealloc)
	}
	if b.formbuff == nil {
		b.formbuff = make([]byte, b.cfg.Body.Form.BufferPrealloc)
	}

	raw, err := b.Bytes()
	if err != nil {
		return nil, err
	}

	switch {
	case mime.Complies(mime.FormUrlencoded, b.request.ContentType):
		f, b.formbuff, err = formdata.ParseURLEncoded(b.form[:0], raw, b.formbuff[:0], b.cfg.URL.Query.DefaultFlagValue)
		return f, err
	case mime.Complies(mime.Multipart, b.request.ContentType):
		boundary, ok := b.multipartBoundary()
		if !ok {
			return nil, status.ErrBadRequest
		}

		return formdata.ParseMultipart(b.cfg, b.form[:0], raw, b.formbuff[:0], boundary)
	default:
		return nil, status.ErrUnsupportedMediaType
	}
}

// Discard discards the rest of the body (if any). If no networking error was encountered,
// nil is returned.
func (b *Body) Discard() error {
	for b.error == nil {
		_, b.error = b.Retrieve()
	}

	if b.error == io.EOF {
		return nil
	}

	return b.error
}

// Error returns a previously encountered error, otherwise nil.
func (b *Body) Error() error {
	return b.error
}

func (b *Body) Reset() error {
	if err := b.Discard(); err != nil {
		return err
	}

	b.error = nil
	b.buff = b.buff[:0]
	return nil
}

func (b *Body) multipartBoundary() (boundary string, ok bool) {
	for key, value := range strutil.WalkKV(strutil.CutParams(b.request.ContentType)) {
		if key == "boundary" {
			if len(boundary) != 0 {
				return "", false
			}

			boundary = value
		}
	}

	return boundary, true
}
