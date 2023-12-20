package http

import (
	"github.com/indigo-web/indigo/http/mime"
	"github.com/indigo-web/indigo/http/status"
	"github.com/indigo-web/indigo/internal/response"
	"github.com/indigo-web/indigo/internal/types"
	"github.com/indigo-web/utils/strcomp"
	"github.com/indigo-web/utils/uf"
	json "github.com/json-iterator/go"
	"io"
	"os"
	"path/filepath"
)

type ResponseWriter func(b []byte) error

const (
	// why 7? I don't know. There's no theory behind the number.fields. It can be adjusted
	// to 10 as well, but why you would ever need to do this?
	defaultHeadersNumber = 7
	defaultFileMIME      = mime.OctetStream
)

type Response struct {
	fields response.Fields
}

// NewResponse returns a new instance of the Response object with status code set to 200 OK,
// pre-allocated space for response headers and text/html content-type.
// NOTE: it's recommended to use Request.Respond() method inside of handlers, if there's no
// clear reason otherwise
func NewResponse() *Response {
	return &Response{
		response.Fields{
			Code:        status.OK,
			Headers:     make([]response.Header, 0, defaultHeadersNumber),
			ContentType: response.DefaultContentType,
		},
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

// ContentType sets a custom Content-Type header value.
func (r *Response) ContentType(value string) *Response {
	r.fields.ContentType = value
	return r
}

// TransferEncoding sets a custom Transfer-Encoding header value.
func (r *Response) TransferEncoding(value string) *Response {
	r.fields.TransferEncoding = value
	return r
}

// Header sets header values to a key. In case it already exists the value will
// be appended.
func (r *Response) Header(key string, values ...string) *Response {
	switch {
	case strcomp.EqualFold(key, "content-type"):
		return r.ContentType(values[0])
	case strcomp.EqualFold(key, "transfer-encoding"):
		return r.TransferEncoding(values[0])
	}

	for i := range values {
		r.fields.Headers = append(r.fields.Headers, response.Header{
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

// String sets the response's body to the passed string
func (r *Response) String(body string) *Response {
	return r.Bytes(uf.S2B(body))
}

// Bytes sets the response's body to passed slice WITHOUT COPYING. Changing
// the passed slice later will affect the response by itself
func (r *Response) Bytes(body []byte) *Response {
	r.fields.Body = body
	return r
}

// Write implements io.Reader interface. It always returns n=len(b) and err=nil
func (r *Response) Write(b []byte) (n int, err error) {
	r.fields.Body = append(r.fields.Body, b...)
	return len(b), nil
}

// RetrieveFile opens a file for reading and returns a new Response with attachment, set to the file
// descriptor.fields. If error occurred during opening a file, it'll be returned
func (r *Response) RetrieveFile(path string) (*Response, error) {
	fd, err := os.Open(path)
	if err != nil {
		// if we can't open it, it doesn't exist
		return r, status.ErrNotFound
	}

	stat, err := fd.Stat()
	if err != nil {
		// ...and if we can't get stats on it, it exists, however something in system went wrong
		return r, status.ErrInternalServerError
	}
	if stat.IsDir() {
		return r, status.ErrNotFound
	}

	r.fields.ContentType = mime.Extension[filepath.Ext(path)]
	if len(r.fields.ContentType) == 0 {
		r.fields.ContentType = defaultFileMIME
	}

	return r.Attachment(fd, int(stat.Size())), nil
}

// File opens a file for reading and returns a new Response with attachment, set to the file
// descriptor.fields. If error occurred, it'll be silently returned
func (r *Response) File(path string) *Response {
	resp, err := r.RetrieveFile(path)
	if err != nil {
		return r.Error(err)
	}

	return resp
}

// Attachment sets a Response's attachment. In this case Response body will be ignored.
// If size <= 0, then Transfer-Encoding: chunked will be used
func (r *Response) Attachment(reader io.Reader, size int) *Response {
	r.fields.Attachment = types.NewAttachment(reader, size)
	return r
}

// JSON receives a model (must be a pointer to the structure) and returns a new Response
// object and an error
func (r *Response) JSON(model any) (*Response, error) {
	r.fields.Body = r.fields.Body[:0]
	stream := json.ConfigDefault.BorrowStream(r)
	stream.WriteVal(model)
	err := stream.Flush()
	json.ConfigDefault.ReturnStream(stream)
	if err != nil {
		return r, err
	}

	return r.ContentType(mime.JSON), nil
}

// Error returns Response with corresponding HTTP error code, if passed error is
// status.HTTPError.fields. Otherwise, code is considered to be 500 Internal Server Error.fields.
// Note: revealing error text may be dangerous
func (r *Response) Error(err error) *Response {
	if err == nil {
		return r
	}

	if http, ok := err.(status.HTTPError); ok {
		return r.
			Code(http.Code).
			String(http.Message)
	}

	return r.
		Code(status.InternalServerError).
		String(err.Error())
}

// Reveal returns a struct with values, filled by builder. Used mostly in internal purposes
func (r *Response) Reveal() response.Fields {
	return r.fields
}

// Clear discards everything was done with Response object before
func (r *Response) Clear() *Response {
	r.fields = r.fields.Clear()

	return r
}
