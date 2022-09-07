package types

import (
	"github.com/fakefloordiv/indigo/http/headers"
	"github.com/fakefloordiv/indigo/http/status"
	"github.com/fakefloordiv/indigo/internal"
)

type (
	ResponseWriter func(b []byte) error
	Render         func(response Response) error
	FileErrHandler func(err error) Response
)

// idk why 5, but why not
const initialRespHeadersSize = 5

// WithResponse is just a nil-filled default pre-created response. Because
// of clear methods, it is anyway copied every time it is used as constructor
// so please, DO NOT modify fields of this variable
var WithResponse = Response{
	Code:   status.OK,
	Status: status.Text(status.OK),
}

type Response struct {
	Code   status.Code
	Status status.Status
	// headers due to possible side effects are decided to be private
	// also uninitialized response must ALWAYS have this value as nil
	headers headers.Headers
	// Body is a mutable object. But it's guaranteed that in WithResponse it will not
	// be modified because it's nil. This means that any data will be appended will
	// allocate a new underlying array
	Body     []byte
	Filename string
	handler  FileErrHandler
}

func NewResponse() Response {
	return Response{
		Code:    status.OK,
		Status:  status.Text(status.OK),
		headers: make(headers.Headers),
	}
}

// WithCode sets a response code and a corresponding status.
// In case of unknown code, "Unknown Status Code" will be set as a status
// code. In this case you should call WithStatus explicitly
func (r Response) WithCode(code status.Code) Response {
	r.Code = code
	r.Status = status.Text(code)
	return r
}

// WithStatus sets a status text. Not compulsory, because http does not force
// us strictly checking a response code, and Unknown Status Code is still a
// valid response code, but you are better to do this. Be a bro
func (r Response) WithStatus(status status.Status) Response {
	r.Status = status
	return r
}

// WithHeader sets header values to a key. In case it already exists it will
// be overridden
func (r Response) WithHeader(key string, values ...string) Response {
	if r.headers == nil {
		r.headers = make(headers.Headers, initialRespHeadersSize)
	}

	r.headers[key] = values

	return r
}

// WithHeaders applies WithHeader to a whole map
func (r Response) WithHeaders(headers headers.Headers) Response {
	response := r

	for key, values := range headers {
		response = response.WithHeader(key, values...)
	}

	return response
}

// WithBody sets a string as a response body. This will override already-existing
// body if it was set
func (r Response) WithBody(body string) Response {
	return r.WithBodyByte(internal.S2B(body))
}

// WithBodyAppend appends a string to already-existing body
func (r Response) WithBodyAppend(body string) Response {
	// constructor is anyway non-clear because of headers, so we do not lose
	// anything
	r.Body = append(r.Body, body...)
	return r
}

// WithBodyByte does all the same as WithBody does, but for byte slices
func (r Response) WithBodyByte(body []byte) Response {
	r.Body = body
	return r
}

// WithBodyByteAppend does all the same as WithBodyAppend does, but with byte slices
func (r Response) WithBodyByteAppend(body []byte) Response {
	r.Body = append(r.Body, body...)
	return r
}

// WithFile sets a file path as a file that is supposed to be uploaded as a
// response. WithFile replaces a response body, so in case last one is specified,
// it'll be ignored.
// In case any error occurred (file not found, or error occurred during reading,
// etc.), handler will be called with a raised error
func (r Response) WithFile(path string, handler FileErrHandler) Response {
	r.Filename = path
	r.handler = handler
	return r
}

// WithError simply sets a code status.InternalServerError and response body
// as an error text
func (r Response) WithError(err error) Response {
	return r.WithCode(status.InternalServerError).WithBody(err.Error())
}

// Headers returns response headers map
func (r Response) Headers() headers.Headers {
	return r.headers
}

// File returns response filename and error handler
func (r Response) File() (string, FileErrHandler) {
	return r.Filename, r.handler
}
