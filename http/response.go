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

// Code sets the response code. If the code is unrecognized, its default status string
// is "Nonstandard". Otherwise, it will be chosen automatically unless overridden.
func (r *Response) Code(code status.Code) *Response {
	r.fields.Code = code
	return r
}

// Status sets a custom status text.
func (r *Response) Status(status status.Status) *Response {
	r.fields.Status = status
	return r
}

// ContentType is a shorthand for Header("Content-Type", value) with an option of setting
// a charset if at least one is specified. All others are ignored.
func (r *Response) ContentType(value mime.MIME, charset ...mime.Charset) *Response {
	if value == mime.Unset {
		return r
	}

	if len(charset) > 0 {
		r.fields.Charset = charset[0]
	}

	return r.Header("Content-Type", value)
}

// Compress chooses and sets the best suiting compression based on client preferences.
func (r *Response) Compress() *Response {
	r.fields.AutoCompress = true
	r.fields.ContentEncoding = "" // to avoid conflicts, wins the last method applied.
	return r
}

// Compression enforces a specific codec to be used, even if it isn't in Accept-Encoding.
// The method is no-op if the token is not recognized.
func (r *Response) Compression(token string) *Response {
	r.fields.ContentEncoding = token
	r.fields.AutoCompress = false
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

// Headers merges the map into the response headers.
func (r *Response) Headers(headers map[string][]string) *Response {
	for k, v := range headers {
		r.Header(k, v...)
	}

	return r
}

// String sets the response body.
func (r *Response) String(body string) *Response {
	return r.Bytes(uf.S2B(body))
}

// Bytes sets the response body. Please note that the passed slice must not be modified
// after being passed.
func (r *Response) Bytes(body []byte) *Response {
	return r.Stream(r.body.Reset(body), int64(len(body)))
}

// Write implements io.Reader interface. It always returns n=len(b) and err=nil
func (r *Response) Write(b []byte) (n int, err error) {
	r.fields.Buffer = append(r.fields.Buffer, b...)
	r.Bytes(r.fields.Buffer)

	return len(b), nil
}

// TryFile tries to open a file by the path for reading and sets it as an upload stream if succeeded.
// Otherwise, the error is returned.
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
		Stream(fd, stat.Size()), nil
}

// File opens a file by the path and sets it as an upload stream if succeeded. Otherwise, the error
// is silently written instead.
func (r *Response) File(path string) *Response {
	resp, err := r.TryFile(path)
	return resp.Error(err)
}

// Stream sets a reader to be the source of the response's body. If no size is provided AND the reader
// doesn't have the Len() int method, the stream is considered unsized and therefore will be streamed
// using chunked transfer encoding. Otherwise, plain transfer is used, unless a compression is applied.
// Specifying the size of -1 forces the stream to be considered unsized.
func (r *Response) Stream(reader io.Reader, size ...int64) *Response {
	type Len interface {
		Len() int
	}

	r.fields.StreamSize = -1
	if len(size) > 0 {
		r.fields.StreamSize = size[0]
	} else if l, ok := reader.(Len); ok {
		r.fields.StreamSize = int64(l.Len())
	}

	r.fields.Stream = reader
	return r
}

// Cookie adds cookies. They'll be later rendered as a set of Set-Cookie headers
func (r *Response) Cookie(cookies ...cookie.Cookie) *Response {
	r.fields.Cookies = append(r.fields.Cookies, cookies...)
	return r
}

// TryJSON tries to serialize the model into JSON.
func (r *Response) TryJSON(model any) (*Response, error) {
	stream := json.ConfigDefault.BorrowStream(r)
	stream.WriteVal(model)
	err := stream.Flush()
	json.ConfigDefault.ReturnStream(stream)

	return r.ContentType(mime.JSON), err
}

// JSON serializes the model into JSON and sets the Content-Type to application/json if succeeded.
// Otherwise, the error is silently written instead.
func (r *Response) JSON(model any) *Response {
	resp, err := r.TryJSON(model)
	return resp.Error(err)
}

// Error returns the response builder with an error set. The nil value for error is a no-op.
// If the error is an instance of status.HTTPError, its status code is used instead the default one.
// The default code is status.ErrInternalServerError, which can be overridden if at least one code is
// specified (all others are ignored).
func (r *Response) Error(err error, code ...status.Code) *Response {
	if err == nil {
		return r
	}

	if http, ok := err.(status.HTTPError); ok {
		return r.Code(http.Code)
	}

	c := status.InternalServerError
	if len(code) > 0 {
		c = code[0]
	}

	return r.
		Code(c).
		String(err.Error())
}

// Buffered allows to enable or disable writes deferring. When enabled, data from body stream
// is read until there is enough space available in an underlying buffer. If the data must be
// flushed soon possible (e.g. polling or proxying), the option should be disabled.
//
// By default, the option is enabled.
func (r *Response) Buffered(flag bool) *Response {
	r.fields.Buffered = flag
	return r
}

// Expose gives direct access to internal builder fields.
func (r *Response) Expose() *response.Fields {
	return &r.fields
}

// Clear discards all changes.
func (r *Response) Clear() *Response {
	r.fields.Clear()
	return r
}

// Respond is a shorthand for request.Respond(). Can be used as a dummy handler.
func Respond(request *Request) *Response {
	return request.Respond()
}

// Code sets the response code. If the code is unrecognized, its default status string
// is "Nonstandard". Otherwise, it will be chosen automatically unless overridden.
func Code(request *Request, code status.Code) *Response {
	return request.Respond().Code(code)
}

// ContentType is a shorthand for request.Respond().ContentType(...)
//
// ContentType itself is a shorthand for Header("Content-Type", value)
// with an option of setting a charset, if at least one is specified. All others are ignored.
func ContentType(request *Request, contentType mime.MIME, charset ...mime.Charset) *Response {
	return request.Respond().ContentType(contentType, charset...)
}

// String sets the response body.
func String(request *Request, str string) *Response {
	return request.Respond().String(str)
}

// Bytes sets the response body. Please note that the passed slice must not be modified
// after being passed.
func Bytes(request *Request, b []byte) *Response {
	return request.Respond().Bytes(b)
}

// File opens a file by the path and sets it as an upload stream if succeeded. Otherwise, the error
// is silently written instead.
func File(request *Request, path string) *Response {
	return request.Respond().File(path)
}

// Stream sets a reader to be the source of the response's body. If no size is provided AND the reader
// doesn't have the Len() int method, the stream is considered unsized and therefore will be streamed
// using chunked transfer encoding. Otherwise, plain transfer is used, unless a compression is applied.
// Specifying the size of -1 forces the stream to be considered unsized.
func Stream(request *Request, reader io.Reader, size ...int64) *Response {
	return request.Respond().Stream(reader, size...)
}

// JSON serializes the model into JSON and sets the Content-Type to application/json if succeeded.
// Otherwise, the error is silently written instead.
func JSON(request *Request, model any) *Response {
	return request.Respond().JSON(model)
}

// Error returns the response builder with an error set. The nil value for error is a no-op.
// If the error is an instance of status.HTTPError, its status code is used instead the default one.
// The default code is status.ErrInternalServerError, which can be overridden if at least one code is
// specified (all others are ignored).
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
