package http

import (
	"io"
	"os"
	"strings"

	"github.com/indigo-web/indigo/http/status"
	"github.com/indigo-web/indigo/internal"
)

type ResponseWriter func(b []byte) error

// IDK why 7, but let it be
const (
	defaultHeadersNumber = 7
	defaultContentType   = "text/html"
)

// Attachment is a wrapper for io.Reader, with the difference that there is the size attribute.
// If positive value (including 0) is set, then ordinary plain-text response will be rendered.
// Otherwise, chunked transfer encoding is used.
type Attachment struct {
	content io.Reader
	size    int
}

// NewAttachment returns a new Attachment instance
func NewAttachment(content io.Reader, size int) Attachment {
	return Attachment{
		content: content,
		size:    size,
	}
}

func (a Attachment) Content() io.Reader {
	return a.content
}

func (a Attachment) Size() int {
	return a.size
}

type Response struct {
	Code status.Code
	// Status is empty by default, in this case renderer must put a default one
	Status status.Status
	// headers are just a slice of strings, length of which is always dividable by 2, because
	// it contains pairs of keys and values
	headers []string
	// ContentType, as a special for core header, should be treated individually
	ContentType string
	// The same is about TransferEncoding
	TransferEncoding string
	Body             []byte

	// attachment is a reader that's going to be read only at response's rendering, so its
	// processing should usually be quite efficient.
	//
	// Note: if attachment is set, Body will be ignored
	attachment Attachment
}

func NewResponse() Response {
	return Response{
		Code:        status.OK,
		headers:     make([]string, 0, defaultHeadersNumber*2),
		ContentType: defaultContentType,
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

// WithContentType sets a custom Content-Type header value.
func (r Response) WithContentType(value string) Response {
	r.ContentType = value
	return r
}

// WithTransferEncoding sets a custom Transfer-Encoding header value.
func (r Response) WithTransferEncoding(value string) Response {
	r.TransferEncoding = value
	return r
}

// WithHeader sets header values to a key. In case it already exists the value will
// be appended.
func (r Response) WithHeader(key string, values ...string) Response {
	switch {
	case strings.EqualFold(key, "content-type"):
		return r.WithContentType(values[0])
	case strings.EqualFold(key, "transfer-encoding"):
		return r.WithTransferEncoding(values[0])
	}

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

// DiscardHeaders returns response object with no any headers set.
//
// Warning: this action is not pure. Appending new headers will cause overriding
// old ones
func (r Response) DiscardHeaders() Response {
	r.headers = r.headers[:0]
	return r
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

// WithWriter takes a function that takes an io.Writer, which allows us to stream data
// directly into the response body.
// Note: this method causes an allocation
//
// TODO: This is not the best design solution. I would like to make this method just like
//
//	all others, so returning only Response object itself. The problem is that it is
//	impossible because io.Writer is a procedure-style thing that does not work with
//	our builder that pretends to be clear. Hope in future this issue will be solved
func (r Response) WithWriter(cb func(io.Writer) error) (Response, error) {
	writer := newBodyIOWriter(r)
	err := cb(writer)

	return writer.response, err
}

// WithFile opens a file for reading, and returns a new response with attachment corresponding
// to the file FD. In case not found or any other error, it'll be directly returned
func (r Response) WithFile(path string) (Response, error) {
	file, err := os.Open(path)
	if err != nil {
		return r, err
	}

	stat, err := file.Stat()
	attachment := NewAttachment(file, int(stat.Size()))

	return r.WithAttachment(attachment), err
}

// WithAttachment sets a response's attachment. In this case response body will be ignored
func (r Response) WithAttachment(attachment Attachment) Response {
	r.attachment = attachment
	return r
}

// WithError tries to set a corresponding status code and response body equal to error text
// if error is known to server, otherwise setting status code to status.InternalServerError
// without setting a response body to the error text, because this possibly may reveal some
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

// Attachment returns response's attachment.
//
// WARNING: do NEVER use this method in your code. It serves internal purposes ONLY
func (r Response) Attachment() Attachment {
	return r.attachment
}

// Clear discards everything was done with response object before
func (r Response) Clear() Response {
	r.Code = status.OK
	r.Status = ""
	r.ContentType = defaultContentType
	r.TransferEncoding = ""
	r.headers = r.headers[:0]
	r.Body = nil
	r.attachment = Attachment{}
	return r
}

// bodyIOWriter is an implementation of io.Writer for response body
type bodyIOWriter struct {
	response Response
	readBuff []byte
}

func newBodyIOWriter(response Response) *bodyIOWriter {
	return &bodyIOWriter{
		response: response,
	}
}

func (r *bodyIOWriter) Write(data []byte) (n int, err error) {
	r.response.Body = append(r.response.Body, data...)

	return len(data), nil
}

func (r *bodyIOWriter) ReadFrom(reader io.Reader) (n int64, err error) {
	const readBuffSize = 2048

	if len(r.readBuff) == 0 {
		r.readBuff = make([]byte, readBuffSize)
	}

	for {
		readN, readErr := reader.Read(r.readBuff)
		_, _ = r.Write(r.readBuff[:n]) // bodyIOWriter.Write always returns n=len(data) and err=nil
		n += int64(readN)

		switch readErr {
		case nil:
		case io.EOF:
			return n, nil
		default:
			return n, readErr
		}
	}
}
