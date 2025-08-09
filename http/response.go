package http

import (
	"io"
	"os"

	"github.com/flrdv/uf"
	"github.com/indigo-web/indigo/http/cookie"
	"github.com/indigo-web/indigo/http/mime"
	"github.com/indigo-web/indigo/http/status"
	"github.com/indigo-web/indigo/internal/response"
	"github.com/indigo-web/indigo/kv"
	json "github.com/json-iterator/go"
)

const (
	// why 7? I honestly don't know. There's no theory nor researches behind this.
	// It can be adjusted to 10 as well, but why would you?
	preallocateResponseHeaders = 7
)

type Response struct {
	body   sliceReader
	fields response.Fields
}

// NewResponse returns a new instance of the Response object with status code set to 200 OK,
// pre-allocated space for response headers and text/html content-type.
// NOTE: it's recommended to use Request.Respond() method inside of handlers, if there's no
// clear reason otherwise
func NewResponse() *Response {
	fields := response.Fields{
		Headers: make([]kv.Pair, 0, preallocateResponseHeaders),
	}
	fields.Clear()

	return &Response{
		body:   sliceReader{},
		fields: fields,
	}
}

// Code sets a Response code and a corresponding status.
// In case of unknown code, "Unknown Status Code" will be set as a status
// code. In this case you should call Status explicitly
func (r *Response) Code(code status.Code) *Response {
	r.fields.Code = code
	return r
}

// Status sets a custom status text. This text does not matter at all, and usually
// totally ignored by client, so there is actually no reasons to use this except some
// rare cases when you need to represent a Response status text somewhere
func (r *Response) Status(status status.Status) *Response {
	r.fields.Status = status
	return r
}

// ContentType is a shorthand for Header("Content-Type", value) with an option of setting
// a charset. If more than 1 is set, only the first one is used.
func (r *Response) ContentType(value mime.MIME, charset ...mime.Charset) *Response {
	if value == mime.Unset {
		return r
	}

	if len(charset) > 0 {
		r.fields.Charset = charset[0]
	}

	return r.Header("Content-Type", value)
}

// Compress sets the Content-Encoding value and compresses the outcoming body. Passing the compression
// token that isn't recognized is a no-op.
func (r *Response) Compress(token string) *Response {
	r.fields.ContentEncoding = token
	return r
}

// Header appends a key-values pair into the list of headers to be sent in the response. Passing
// Content-Encoding isn't equivalent to calling Compress() and ultimately results in no encodings
// being automatically applied. Can be used in order to use own compressors.
func (r *Response) Header(key string, values ...string) *Response {
	for i := range values {
		r.fields.Headers = append(r.fields.Headers, kv.Pair{
			Key:   key,
			Value: values[i],
		})
	}

	return r
}

// Headers simply merges passed headers into Response. Also, it is the only
// way to specify a quality marker of value. In case headers were not initialized
// before, Response headers will be set to a passed map, so editing this map
// will affect Response
func (r *Response) Headers(headers map[string][]string) *Response {
	resp := r

	for k, v := range headers {
		resp = resp.Header(k, v...)
	}

	return resp
}

// String sets the response body.
func (r *Response) String(body string) *Response {
	return r.Bytes(uf.S2B(body))
}

// Bytes sets the response body without copying it.
func (r *Response) Bytes(body []byte) *Response {
	return r.SizedStream(r.body.Reset(body), int64(len(body)))
}

// Write implements io.Reader interface. It always returns n=len(b) and err=nil
func (r *Response) Write(b []byte) (n int, err error) {
	r.fields.Buffer = append(r.fields.Buffer, b...)
	r.Bytes(r.fields.Buffer)

	return len(b), nil
}

// TryFile tries to open a file for reading and returns a new Response with attachment.
func (r *Response) TryFile(path string) (*Response, error) {
	fd, err := os.Open(path)
	if err != nil {
		// if we can't open it, it doesn't exist
		return r, status.ErrNotFound
	}

	stat, err := fd.Stat()
	if err != nil {
		return r, status.ErrInternalServerError
	}
	if stat.IsDir() {
		return r, status.ErrNotFound
	}

	return r.
		ContentType(mime.Guess(path, mime.HTML)).
		SizedStream(fd, stat.Size()), nil
}

// File opens a file for reading and returns a new Response with attachment, set to the file
// descriptor.fields. If error occurred, it'll be silently returned
func (r *Response) File(path string) *Response {
	resp, err := r.TryFile(path)
	if err != nil {
		return r.Error(err)
	}

	return resp
}

// Stream sets a reader to be the source of the response's body.
func (r *Response) Stream(reader io.Reader) *Response {
	// TODO: we can check whether the reader implements Len() int interface and in that
	// TODO: case elide the chunked transfer encoding
	r.fields.Stream = reader
	r.fields.StreamSize = -1
	return r
}

// SizedStream receives a hint of the stream's future size. This helps, for example, uploading files,
// as in this case we can rely on io.WriterTo interface, which might use more effective kernel mechanisms
// available, e.g. sendfile(2) for Linux. Passing the size of -1 is effectively equivalent to just Stream().
func (r *Response) SizedStream(reader io.Reader, size int64) *Response {
	r.fields.Stream = reader
	r.fields.StreamSize = size
	return r
}

// Cookie adds cookies. They'll be later rendered as a set of Set-Cookie headers
func (r *Response) Cookie(cookies ...cookie.Cookie) *Response {
	r.fields.Cookies = append(r.fields.Cookies, cookies...)
	return r
}

// TryJSON receives a model (must be a pointer to the structure) and returns a new Response
// object and an error
func (r *Response) TryJSON(model any) (*Response, error) {
	stream := json.ConfigDefault.BorrowStream(r)
	stream.WriteVal(model)
	err := stream.Flush()
	json.ConfigDefault.ReturnStream(stream)

	return r.ContentType(mime.JSON), err
}

// JSON does the same as TryJSON does, except returned error is being implicitly wrapped
// by Error
func (r *Response) JSON(model any) *Response {
	resp, err := r.TryJSON(model)
	if err != nil {
		return r.Error(err)
	}

	return resp
}

// Error returns a response builder with an error set. If passed err is nil, nothing will happen.
// If an instance of status.HTTPError is passed, error code will be automatically set. Custom
// codes can be passed, however only first will be used. By default, the error is
// status.ErrInternalServerError
func (r *Response) Error(err error, code ...status.Code) *Response {
	if err == nil {
		return r
	}

	if http, ok := err.(status.HTTPError); ok {
		return r.Code(http.Code)
	}

	c := status.InternalServerError
	if len(code) > 0 {
		// peek the first, ignore the rest
		c = code[0]
	}

	return r.
		Code(c).
		String(err.Error())
}

// Expose gives direct access to internal builder fields.
func (r *Response) Expose() *response.Fields {
	return &r.fields
}

// Clear discards everything was done with Response object before.
func (r *Response) Clear() *Response {
	r.fields.Clear()
	return r
}

// Respond is a shorthand for request.Respond(). May be used as a dummy handler.
func Respond(request *Request) *Response {
	return request.Respond()
}

// Code is a shorthand for request.Respond().Code(...)
func Code(request *Request, code status.Code) *Response {
	return request.Respond().Code(code)
}

// String is a shorthand for request.Respond().String(...)
func String(request *Request, str string) *Response {
	return request.Respond().String(str)
}

// Bytes is a shorthand for request.Respond().Bytes(...)
func Bytes(request *Request, b []byte) *Response {
	return request.Respond().Bytes(b)
}

// File is a shorthand for request.Respond().File(...)
func File(request *Request, path string) *Response {
	return request.Respond().File(path)
}

// Stream is a shorthand for request.Respond().Stream(...)
func Stream(request *Request, reader io.Reader) *Response {
	return request.Respond().Stream(reader)
}

// SizedStream is a shorthand for request.Respond().SizedStream(...)
func SizedStream(request *Request, reader io.Reader, size int64) *Response {
	return request.Respond().SizedStream(reader, size)
}

// JSON is a shorthand for request.Respond().JSON(...)
func JSON(request *Request, model any) *Response {
	return request.Respond().JSON(model)
}

// Error is a shorthand for request.Respond().Error(...)
//
// Error returns the response builder with an error set. If passed err is nil, nothing will happen.
// If an instance of status.HTTPError is passed, its status code is automatically set. Otherwise,
// status.ErrInternalServerError is used. A custom code can be set. Passing multiple status codes
// will discard all except the first one.
func Error(request *Request, err error, code ...status.Code) *Response {
	return request.Respond().Error(err, code...)
}

type sliceReader struct {
	data []byte
}

func (s *sliceReader) Read(b []byte) (n int, err error) {
	n = copy(b, s.data)
	s.data = s.data[n:]
	if len(s.data) == 0 {
		err = io.EOF
	}

	return n, err
}

func (s *sliceReader) Reset(data []byte) *sliceReader {
	s.data = data
	return s
}
