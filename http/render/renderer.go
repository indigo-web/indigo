package render

import (
	"errors"
	"github.com/fakefloordiv/indigo/http"
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
	buff      []byte
	keepAlive bool

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
// 7) Handle closing connection in case protocol requires or has not specified the opposite
func (r *Renderer) Response(
	request *types.Request, response types.Response, writer types.ResponseWriter,
) (err error) {
	if !r.keepAlive {
		// in case this value is false, this can mean only 1 thing - it is not initialized
		// and once it is initialized, it is supposed to be true, otherwise connection will
		// be closed after this call anyway
		r.keepAlive = isKeepAlive(request)
		if !r.keepAlive {
			err = http.ErrCloseConnection
		}
	}

	buff := r.buff[:0]
	buff = append(append(buff, proto.ToBytes(request.Proto)...), space...)
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

			return r.Response(request, handler(err), writer)
		}

		return err
	}

	buff = renderContentLength(len(response.Body), buff)
	r.buff = append(append(buff, crlf...), response.Body...)

	writerErr := writer(r.buff)
	if writerErr != nil {
		err = writerErr
	}

	return err
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

func renderHeader(key, value string, into []byte) []byte {
	return append(append(append(into, key...), colonSpace...), value...)
}

func isKeepAlive(request *types.Request) bool {
	if request.Proto == proto.HTTP09 {
		return false
	}

	keepAlive, found := request.Headers["connection"]
	switch found {
	case true:
		return keepAlive == "keep-alive"
	case false:
		switch request.Proto {
		case proto.HTTP10:
			// by default http/1.0 is not keep-alive. To be, it must
			// specify it explicitly
			return false
		case proto.HTTP11:
			return true
		}
	}

	return false
}
