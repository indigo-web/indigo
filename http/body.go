package http

import (
	"github.com/indigo-web/indigo/config"
	"github.com/indigo-web/indigo/http/mime"
	"github.com/indigo-web/indigo/http/status"
	"github.com/indigo-web/indigo/internal/keyvalue"
	"github.com/indigo-web/indigo/internal/qparams"
	"github.com/indigo-web/utils/uf"
	json "github.com/json-iterator/go"
	"io"
)

type BodyCallback func([]byte) error

type Retriever interface {
	// Retrieve reads and returns a piece of body available for processing
	Retrieve() ([]byte, error)
	Init(*Request)
}

type FormData = keyvalue.Storage

// a hack to embed the retriever privately
type retriever = Retriever

type Body struct {
	retriever
	request  *Request
	formData *FormData
	cfg      *config.Config
	error    error
	buff     []byte
	pending  []byte
}

func NewBody(impl retriever, cfg *config.Config) *Body {
	return &Body{
		retriever: impl,
		formData:  keyvalue.New(),
		cfg:       cfg,
	}
}

// Callback invokes the callback every time as there's a piece of body available
// for reading. If the callback returns an error, it'll be passed back to the caller.
// The callback is not notified when there's no more data or networking error has
// occurred.
//
// Please note: this method can't be called more than once.
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
		b.buff = make([]byte, b.cfg.Body.BufferPrealloc)
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
// with mime.JSON (in this case, status.ErrUnsupportedMediaType is returned). It also
// can't be called more than once.
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
func (b *Body) Form() (*FormData, error) {
	if !mime.Complies(mime.FormUrlencoded, b.request.ContentType) {
		return b.formData, status.ErrUnsupportedMediaType
	}

	raw, err := b.Bytes()
	if err != nil {
		return b.formData, err
	}

	return b.formData, qparams.Parse(raw, qparams.Into(b.formData))
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

// Init MUST NOT be used, as it may cause deadlock. FOR INTERNAL PURPOSES ONLY.
func (b *Body) Init(r *Request) {
	b.formData.Clear()
	// TODO: there must be a better way to solve the circular referencing problem
	b.request = r
	b.error = nil
	b.buff = b.buff[:0]
	b.retriever.Init(r)
}
