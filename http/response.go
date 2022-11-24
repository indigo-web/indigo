package http

import (
	"github.com/fakefloordiv/indigo/http/status"
	"github.com/fakefloordiv/indigo/internal"
)

type (
	ResponseWriter func(b []byte) error
	Render         func(response Response) error
	FileErrHandler func(err error) Response
)

// idk why 7, but let it be
const defaultHeadersNumber = 7

type Response struct {
	Code status.Code
	// Status is empty by default, in this case renderer must put a default one
	Status status.Status
	// headers are just a slice of strings, length of which is always dividable by 2, because
	// it contains pairs of keys and values
	headers []string
	// Body is a mutable object. But it's guaranteed that in Response it will not
	// be modified because it's nil. This means that any data will be appended will
	// allocate a new underlying array
	Body     []byte
	Filename string
	handler  FileErrHandler
}

func NewResponse() Response {
	return Response{
		Code:    status.OK,
		headers: make([]string, 0, defaultHeadersNumber*2),
	}
}

// WithCode sets a response code and a corresponding status.
// In case of unknown code, "Unknown Status Code" will be set as a status
// code. In this case you should call Status explicitly
func (r Response) WithCode(code status.Code) Response {
	r.Code = code
	return r
}

// WithStatus sets a custom status text. This text does not matter at all, and usually
// totally ignored by client, so there is actually no reasons to use this except some
// rare cases when you need to represent a response status text somewhere
func (r Response) WithStatus(status status.Status) Response {
	r.Status = status
	return r
}

// WithHeader sets header values to a key. In case it already exists the value will
// be appended
func (r Response) WithHeader(key string, values ...string) Response {
	for i := range values {
		r.headers = append(r.headers, key, values[i])
	}

	return r
}

// WithHeaders simply merges passed headers into response. Also, it is the only
// way to specify a quality marker of value. In case headers were not initialized
// before, response headers will be set to a passed map, so editing this map
// will affect response
func (r Response) WithHeaders(headers map[string][]string) Response {
	resp := r

	for k, v := range headers {
		resp = resp.WithHeader(k, v...)
	}

	return resp
}

// WithBody sets a string as a response body. This will override already-existing
// body if it was set
func (r Response) WithBody(body string) Response {
	return r.WithBodyByte(internal.S2B(body))
}

// WithBodyByte does all the same as Body does, but for byte slices
func (r Response) WithBodyByte(body []byte) Response {
	r.Body = body
	return r
}

// WithFile sets a file path as a file that is supposed to be uploaded as a
// response. File replaces a response body, so in case last one is specified,
// it'll be ignored.
// In case any error occurred (file not found, or error occurred during reading,
// etc.), handler will be called with a raised error
func (r Response) WithFile(path string, handler FileErrHandler) Response {
	r.Filename = path
	r.handler = handler
	return r
}

// WithError tries to set a corresponding status code and response body equal to error text
// if error is known to server, otherwise setting status code to status.InternalServerError
// without setting a response body to the error text, because this possibly can reveal some
// sensitive internal infrastructure details
func (r Response) WithError(err error) Response {
	resp := r.WithBody(err.Error())

	switch err {
	case status.ErrBadRequest:
		return resp.WithCode(status.BadRequest)
	case status.ErrNotFound:
		return resp.WithCode(status.NotFound)
	case status.ErrMethodNotAllowed:
		return resp.WithCode(status.MethodNotAllowed)
	case status.ErrTooLarge, status.ErrURITooLong:
		return resp.WithCode(status.RequestEntityTooLarge)
	case status.ErrHeaderFieldsTooLarge:
		return resp.WithCode(status.RequestHeaderFieldsTooLarge)
	case status.ErrUnsupportedProtocol:
		return resp.WithCode(status.NotImplemented)
	case status.ErrUnsupportedEncoding:
		return resp.WithCode(status.NotAcceptable)
	case status.ErrConnectionTimeout:
		return resp.WithCode(status.RequestTimeout)
	default:
		return r.WithCode(status.InternalServerError)
	}
}

// Headers returns an underlying response headers
func (r Response) Headers() []string {
	return r.headers
}

// File returns response filename and error handler
func (r Response) File() (string, FileErrHandler) {
	return r.Filename, r.handler
}

func (r Response) Reset() {
	r.Code = status.OK
	r.Status = ""
	r.headers = r.headers[:0]
	r.Filename = ""
	r.Body = nil
	r.handler = nil
}
