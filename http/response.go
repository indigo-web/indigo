package http

import (
	"github.com/indigo-web/indigo/internal/render/types"
	json "github.com/json-iterator/go"
	"io"
	"os"
	"strings"

	"github.com/indigo-web/indigo/http/status"
	"github.com/indigo-web/utils/uf"
)

type ResponseWriter func(b []byte) error

// IDK why 7, but let it be
const (
	defaultHeadersNumber = 7
	defaultContentType   = "text/html"
)

type Response struct {
	attachment       types.Attachment
	Status           status.Status
	ContentType      string
	TransferEncoding string
	headers          []string
	Body             []byte
	Code             status.Code
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
	return r.WithBodyByte(uf.S2B(body))
}

// WithBodyByte does all the same as Body does, but for byte slices
func (r Response) WithBodyByte(body []byte) Response {
	r.Body = body
	return r
}

// WithWriter takes a function with io.Writer receiver. This writer is the actual writer to the response
// body. The code accessing the writer must be wrapped into the function, as the Response builder is
// pretty limited in such a things. It pretends to be clear (all the methods has by-value receivers)
// in order to enable calls chaining, so it's pretty difficult to handle with being io.Writer-compatible
//
// Note: returned error is ALWAYS the error returned by callback. So it may be ignored in cases, when
// callback constantly returns nil
func (r Response) WithWriter(cb func(io.Writer) error) (Response, error) {
	writer := newBodyIOWriter(r)
	err := cb(writer)

	return writer.response, err
}

// WithFile opens a file for reading, and returns a new response with attachment corresponding
// to the file FD. In case not found or any other error, it'll be directly returned.
// In case error occurred while opening the file, response builder won't be affected and stay
// clean
func (r Response) WithFile(path string) (Response, error) {
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

// WithAttachment sets a response's attachment. In this case response body will be ignored.
// If size <= 0, then Transfer-Encoding: chunked will be used
func (r Response) WithAttachment(reader io.Reader, size int) Response {
	r.attachment = types.NewAttachment(reader, size)
	return r
}

// WithJSON receives a model (must be a pointer to the structure) and returns a new Response
// object and an error
func (r Response) WithJSON(model any) (Response, error) {
	r.Body = r.Body[:0]

	resp, err := r.WithWriter(func(w io.Writer) error {
		stream := json.ConfigDefault.BorrowStream(w)
		stream.WriteVal(model)
		err := stream.Flush()
		json.ConfigDefault.ReturnStream(stream)

		return err
	})

	if err != nil {
		return r, err
	}

	return resp.WithContentType("application/json"), nil
}

// WithError checks, whether the passed error is a HTTPError instance. In this case,
// setting response code and body to HTTPError.Code and HTTPError.Message respectively.
// If the check failed, simply setting the code to status.InternalServerError. Error
// message won't be included in the response, as this possibly can spoil project internals,
// creating security breaches
func (r Response) WithError(err error) Response {
	if http, ok := err.(status.HTTPError); ok {
		return r.
			WithCode(http.Code).
			WithBody(http.Message)
	}

	return r.WithCode(status.InternalServerError)
}

// Headers returns an underlying response headers
func (r Response) Headers() []string {
	return r.headers
}

// Attachment returns response's attachment.
//
// WARNING: do NEVER use this method in your code. It serves internal purposes ONLY
func (r Response) Attachment() types.Attachment {
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
	r.attachment = types.Attachment{}
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
