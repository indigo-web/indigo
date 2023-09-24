package http

import (
	"github.com/indigo-web/indigo/http/status"
	"github.com/indigo-web/indigo/internal/render/types"
	"github.com/indigo-web/utils/strcomp"
	"github.com/indigo-web/utils/uf"
	json "github.com/json-iterator/go"
	"io"
	"os"
)

type ResponseWriter func(b []byte) error

// IDK why 7, but let it be
const (
	defaultHeadersNumber = 7
	defaultContentType   = "text/html"
)

type Response struct {
	attachment  types.Attachment
	Status      status.Status
	ContentType string
	// TODO: add corresponding Content-Encoding field
	// TODO: automatically apply the encoding on a body when specified
	TransferEncoding string
	headers          []string
	Body             []byte
	Code             status.Code
}

func NewResponse() *Response {
	return &Response{
		Code:        status.OK,
		headers:     make([]string, 0, defaultHeadersNumber*2),
		ContentType: defaultContentType,
	}
}

// WithCode sets a Response code and a corresponding status.
// In case of unknown code, "Unknown Status Code" will be set as a status
// code. In this case you should call Status explicitly
func (r *Response) WithCode(code status.Code) *Response {
	r.Code = code
	return r
}

// WithStatus sets a custom status text. This text does not matter at all, and usually
// totally ignored by client, so there is actually no reasons to use this except some
// rare cases when you need to represent a Response status text somewhere
func (r *Response) WithStatus(status status.Status) *Response {
	r.Status = status
	return r
}

// WithContentType sets a custom Content-Type header value.
func (r *Response) WithContentType(value string) *Response {
	r.ContentType = value
	return r
}

// WithTransferEncoding sets a custom Transfer-Encoding header value.
func (r *Response) WithTransferEncoding(value string) *Response {
	r.TransferEncoding = value
	return r
}

// WithHeader sets header values to a key. In case it already exists the value will
// be appended.
func (r *Response) WithHeader(key string, values ...string) *Response {
	switch {
	case strcomp.EqualFold(key, "content-type"):
		return r.WithContentType(values[0])
	case strcomp.EqualFold(key, "transfer-encoding"):
		return r.WithTransferEncoding(values[0])
	}

	for i := range values {
		r.headers = append(r.headers, key, values[i])
	}

	return r
}

// WithHeaders simply merges passed headers into Response. Also, it is the only
// way to specify a quality marker of value. In case headers were not initialized
// before, Response headers will be set to a passed map, so editing this map
// will affect Response
func (r *Response) WithHeaders(headers map[string][]string) *Response {
	resp := r

	for k, v := range headers {
		resp = resp.WithHeader(k, v...)
	}

	return resp
}

// WithBody sets a string as a Response body. This will override already-existing
// body if it was set
func (r *Response) WithBody(body string) *Response {
	return r.WithBodyByte(uf.S2B(body))
}

// WithBodyByte does all the same as Body does, but for byte slices
func (r *Response) WithBodyByte(body []byte) *Response {
	r.Body = body
	return r
}

// Write implements io.Reader interface. It always returns n=len(b) and err=nil
func (r *Response) Write(b []byte) (n int, err error) {
	r.Body = append(r.Body, b...)
	return len(b), nil
}

// WithFile opens a file for reading, and returns a new Response with attachment corresponding
// to the file FD. In case not found or any other error, it'll be directly returned.
// In case error occurred while opening the file, Response builder won't be affected and stay
// clean
func (r *Response) WithFile(path string) (*Response, error) {
	file, err := os.Open(path)
	if err != nil {
		return r, err
	}

	stat, err := file.Stat()
	if err != nil {
		return r, err
	}

	return r.WithAttachment(file, int(stat.Size())), nil
}

// WithAttachment sets a Response's attachment. In this case Response body will be ignored.
// If size <= 0, then Transfer-Encoding: chunked will be used
func (r *Response) WithAttachment(reader io.Reader, size int) *Response {
	r.attachment = types.NewAttachment(reader, size)
	return r
}

// WithJSON receives a model (must be a pointer to the structure) and returns a new Response
// object and an error
func (r *Response) WithJSON(model any) (*Response, error) {
	r.Body = r.Body[:0]

	stream := json.ConfigDefault.BorrowStream(r)
	stream.WriteVal(model)
	err := stream.Flush()
	json.ConfigDefault.ReturnStream(stream)
	if err != nil {
		return r, err
	}

	return r.WithContentType("application/json"), nil
}

// WithError returns Response with corresponding HTTP error code, if passed error is
// status.HTTPError. Otherwise, code is considered to be 500 Internal Server Error.
// Note: revealing error text may be dangerous
func (r *Response) WithError(err error) *Response {
	if http, ok := err.(status.HTTPError); ok {
		return r.
			WithCode(http.Code).
			WithBody(http.Message)
	}

	return r.
		WithCode(status.InternalServerError).
		WithBody(err.Error())
}

// Headers returns an underlying Response headers
func (r *Response) Headers() []string {
	return r.headers
}

// Attachment returns Response's attachment.
//
// Note: it serves mostly internal purposes
func (r *Response) Attachment() types.Attachment {
	return r.attachment
}

// Clear discards everything was done with Response object before
func (r *Response) Clear() *Response {
	r.Code = status.OK
	r.Status = ""
	r.ContentType = defaultContentType
	r.TransferEncoding = ""
	r.headers = r.headers[:0]
	r.Body = nil
	r.attachment = types.Attachment{}
	return r
}
