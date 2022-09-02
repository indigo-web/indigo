package render

import (
	"errors"
	"io"
	"math"
	"os"
	"strconv"

	"github.com/fakefloordiv/indigo/http/headers"
	"github.com/fakefloordiv/indigo/http/proto"
	"github.com/fakefloordiv/indigo/http/status"
	"github.com/fakefloordiv/indigo/types"
)

var (
	space         = []byte(" ")
	crlf          = []byte("\r\n")
	colonSpace    = []byte(": ")
	contentLength = []byte("Content-Length: ")

	errConnWrite = errors.New("error occurred while communicating with conn")
)

// Renderer is a session responses renderer. Its purpose is only to know
// something about client, and knowing them, render correct response
// for example, we SHOULD not use content-codings for HTTP/1.0 clients,
// and MUST NOT use them for HTTP/0.9 clients
// Also in case of file is being sent, it collects some meta about it,
// and compresses using available on both server and client encoders
type Renderer struct {
	buff []byte

	defaultHeaders headers.Headers
}

func NewRenderer(buff []byte) *Renderer {
	return &Renderer{
		buff: buff,
	}
}

func (r *Renderer) SetDefaultHeaders(headers headers.Headers) {
	r.defaultHeaders = headers
}

// Response method is rendering types.Response object into some buffer and then writes
// it into the writer. Response method must provide next functionality:
// 1) Render types.Response object according to the provided protocol version
// 2) Be sure that used features of response are supported by client (provided protocol)
// 3) Support default headers
// 4) Add system-important headers, e.g. Content-Length
// 5) Content encodings must be applied here
// 6) Stream-based files uploading must be supported
func (r *Renderer) Response(
	protocol proto.Proto, response types.Response, writer types.ResponseWriter,
) error {
	buff := r.buff[:0]
	buff = append(append(buff, proto.ToBytes(protocol)...), space...)
	buff = append(append(buff, strconv.Itoa(int(response.Code))...), space...)
	buff = append(append(buff, status.Text(response.Code)...), crlf...)

	reqHeaders := response.Headers()

	for key, value := range reqHeaders {
		buff = append(renderHeader(key, value, buff), crlf...)
	}

	for key, value := range r.defaultHeaders {
		_, found := reqHeaders[key]
		if !found {
			buff = append(renderHeader(key, value, buff), crlf...)
		}
	}

	if len(response.Filename) > 0 {
		r.buff = buff

		switch err := r.renderFileInto(writer, response); err {
		case nil:
		case errConnWrite:
			return err
		default:
			_, handler := response.File()

			return r.Response(protocol, handler(err), writer)
		}

		return nil
	}

	buff = renderContentLength(len(response.Body), buff)
	r.buff = append(append(buff, crlf...), response.Body...)

	return writer(r.buff)
}

// renderFileInto opens a file in os.O_RDONLY mode, reading its size and appends
// a Content-Length header equal to size of the file, after that headers are being
// sent. Then 64kb buffer is allocated for reading from file and writing to the
// connection. In case network error occurred, errConnWrite is returned. Otherwise,
// received error is returned
//
// Not very elegant solution, but uploading files is not the main purpose of web-server.
// For small and medium projects, this may be enough, for anything serious - most of all
// nginx will be used (the same is about https)
func (r *Renderer) renderFileInto(writer types.ResponseWriter, response types.Response) error {
	file, err := os.OpenFile(response.Filename, os.O_RDONLY, 69420) // anyway unused
	if err != nil {
		return err
	}

	stat, err := file.Stat()
	if err != nil {
		return err
	}

	r.buff = renderContentLength(int(stat.Size()), r.buff)

	if err = writer(append(r.buff, crlf...)); err != nil {
		return errConnWrite
	}

	// write by blocks 64kb each
	buff := make([]byte, math.MaxUint16)

	for {
		n, err := file.Read(buff)
		switch err {
		case nil:
		case io.EOF:
			return nil
		default:
			return err
		}

		if err = writer(buff[:n]); err != nil {
			return errConnWrite
		}
	}
}

func renderContentLength(value int, buff []byte) []byte {
	return append(append(append(buff, contentLength...), strconv.Itoa(value)...), crlf...)
}

func renderHeader(key string, value []byte, into []byte) []byte {
	return append(append(append(into, key...), colonSpace...), value...)
}
