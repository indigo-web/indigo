package types

import (
	"github.com/fakefloordiv/indigo/http"
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

// WithHeader sets header values to a key. In case it already exists the value will
// be appended
func (r Response) WithHeader(key string, values ...string) Response {
	if r.headers == nil {
		r.headers = make(headers.Headers, initialRespHeadersSize)
	}

	hdrs, found := r.headers[key]
	if !found {
		r.headers[key] = strHeaders2Headers(values...)
	} else {
		r.headers[key] = append(hdrs, strHeaders2Headers(values...)...)
	}

	return r
}

// WithHeaderQ appends or creates a new response header value with a specified
// quality-marker
func (r Response) WithHeaderQ(key string, value string, q uint8) Response {
	if r.headers == nil {
		r.headers = make(headers.Headers, initialRespHeadersSize)
		r.headers[key] = []headers.Header{
			{
				Value: value,
				Q:     q,
			},
		}

		return r
	}

	hdrs, found := r.headers[key]
	if !found {
		r.headers[key] = []headers.Header{
			{
				Value: value,
				Q:     q,
			},
		}
	} else {
		r.headers[key] = append(hdrs, headers.Header{
			Value: value,
			Q:     q,
		})
	}

	return r
}

// WithHeaders simply merges passed headers into response. Also, it is the only
// way to specify a quality marker of value. In case headers were not initialized
// before, response headers will be set to a passed map, so editing this map
// will affect response
func (r Response) WithHeaders(headers headers.Headers) Response {
	if r.headers == nil {
		r.headers = headers
		return r
	}

	for k, v := range headers {
		r.headers[k] = v
	}

	return r
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
	resp := r.WithBody(err.Error())

	switch err {
	case http.ErrBadRequest:
		return resp.WithCode(status.BadRequest)
	case http.ErrNotFound:
		return resp.WithCode(status.NotFound)
	case http.ErrMethodNotAllowed:
		return resp.WithCode(status.MethodNotAllowed)
	case http.ErrTooLarge, http.ErrURITooLong:
		return resp.WithCode(status.RequestEntityTooLarge)
	case http.ErrHeaderFieldsTooLarge:
		return resp.WithCode(status.RequestHeaderFieldsTooLarge)
	case http.ErrUnsupportedProtocol:
		return resp.WithCode(status.NotImplemented)
	case http.ErrUnsupportedEncoding:
		return resp.WithCode(status.NotAcceptable)
	case http.ErrConnectionTimeout:
		return resp.WithCode(status.RequestTimeout)
	default:
		// failed to determine actual error, most of all this is some
		// user's error, so 500 Internal Server Error is good here
		return r.WithCode(status.InternalServerError)
	}
}

// Headers returns response headers map
func (r Response) Headers() headers.Headers {
	return r.headers
}

// File returns response filename and error handler
func (r Response) File() (string, FileErrHandler) {
	return r.Filename, r.handler
}

func strHeaders2Headers(strHeaders ...string) []headers.Header {
	hdrs := make([]headers.Header, len(strHeaders))

	for i := range strHeaders {
		hdrs[i] = headers.Header{
			Value: strHeaders[i],
		}
	}

	return hdrs
}

// OK returns a 200 OK response
func OK() Response {
	return WithResponse
}

// WithCode sets a response code and a corresponding status.
// In case of unknown code, "Unknown Status Code" will be set as a status
// code. In this case you should call WithStatus explicitly
func WithCode(code status.Code) Response {
	return WithResponse.WithCode(code)
}

// WithStatus sets a status text. Not compulsory, because http does not force
// us strictly checking a response code, and Unknown Status Code is still a
// valid response code, but you are better to do this. Be bro
func WithStatus(status status.Status) Response {
	return WithResponse.WithStatus(status)
}

// WithHeader sets header values to a key. In case it already exists the value will
// be appended
func WithHeader(key string, values ...string) Response {
	return WithResponse.WithHeader(key, values...)
}

// WithHeaderQ appends or creates a new response header value with a specified
// quality-marker
func WithHeaderQ(key string, value string, q uint8) Response {
	return WithResponse.WithHeaderQ(key, value, q)
}

// WithHeaders simply merges passed headers into response. Also, it is the only
// way to specify a quality marker of value. In case headers were not initialized
// before, response headers will be set to a passed map, so editing this map
// will affect response
func WithHeaders(headers headers.Headers) Response {
	return WithResponse.WithHeaders(headers)
}

// WithBody sets a string as a response body. This will override already-existing
// body if it was set
func WithBody(body string) Response {
	return WithResponse.WithBody(body)
}

// WithBodyAppend appends a string to already-existing body
func WithBodyAppend(body string) Response {
	return WithResponse.WithBodyAppend(body)
}

// WithBodyByte does all the same as WithBody does, but for byte slices
func WithBodyByte(body []byte) Response {
	return WithResponse.WithBodyByte(body)
}

// WithBodyByteAppend does all the same as WithBodyAppend does, but with byte slices
func WithBodyByteAppend(body []byte) Response {
	return WithResponse.WithBodyByteAppend(body)
}

// WithFile sets a file path as a file that is supposed to be uploaded as a
// response. WithFile replaces a response body, so in case last one is specified,
// it'll be ignored.
// In case any error occurred (file not found, or error occurred during reading,
// etc.), handler will be called with a raised error
func WithFile(path string, handler FileErrHandler) Response {
	return WithResponse.WithFile(path, handler)
}

// WithError simply sets a code status.InternalServerError and response body
// as an error text
func WithError(err error) Response {
	return WithResponse.WithError(err)
}
