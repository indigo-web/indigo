package render

import (
	"errors"
	"io"
	"math"
	"os"
	"strconv"

	methods "github.com/fakefloordiv/indigo/http/method"

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

func NewRenderer(buff []byte, defaultHeaders headers.Headers) *Renderer {
	return &Renderer{
		buff:           buff,
		defaultHeaders: defaultHeaders,
	}
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
	switch r.keepAlive {
	case false:
		// in case this value is false, this can mean only 1 thing - it is not initialized
		// and once it is initialized, it is supposed to be true, otherwise connection will
		// be closed after this call anyway
		r.keepAlive = isKeepAlive(request)
		if !r.keepAlive {
			err = http.ErrCloseConnection
		}
	case true:
		// in case request Connection header is set to close, this response must be the last
		// one, after which one connection will be closed. It's better to close it silently
		connection, found := request.Headers["connection"]
		if found && connection[0].Value == "close" {
			err = http.ErrCloseConnection
		}
	}

	buff := r.buff[:0]
	buff = append(append(buff, proto.ToBytes(request.Proto)...), space...)
	buff = append(append(buff, strconv.Itoa(int(response.Code))...), space...)
	buff = append(append(buff, status.Text(response.Code)...), crlf...)

	customRespHeaders := response.Headers()
	// TODO: this shit decreses performance from 75-77k rps to 68-70k
	respHeaders := mergeHeaders(r.defaultHeaders, customRespHeaders)

	for key, values := range respHeaders {
		buff = append(renderHeader(key, values, buff), crlf...)
	}

	if len(response.Filename) > 0 {
		r.buff = buff

		switch err := r.renderFileInto(request.Method, writer, response); err {
		case nil:
		case errConnWrite:
			return err
		default:
			_, handler := response.File()

			return r.Response(request, handler(err), writer)
		}

		return err
	}

	buff = renderContentLength(int64(len(response.Body)), buff)
	r.buff = append(buff, crlf...)

	// HEAD requests MUST NOT contain message body - the main difference
	// between HEAD and GET requests
	// See rfc2068 9.4
	if request.Method != methods.HEAD {
		r.buff = append(r.buff, response.Body...)
	}

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
func (r *Renderer) renderFileInto(
	method methods.Method, writer types.ResponseWriter, response types.Response,
) error {
	file, err := os.OpenFile(response.Filename, os.O_RDONLY, 0)
	if err != nil {
		return err
	}

	stat, err := file.Stat()
	if err != nil {
		return err
	}

	r.buff = renderContentLength(stat.Size(), r.buff)

	if err = writer(append(r.buff, crlf...)); err != nil {
		return errConnWrite
	}

	if method == methods.HEAD {
		// once again, HEAD requests MUST NOT contain response bodies. They are just like
		// GET request, but without response entities
		return nil
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

func renderContentLength(value int64, buff []byte) []byte {
	buff = append(buff, contentLength...)

	return append(strconv.AppendInt(buff, value, 10), crlf...)
}

func renderHeader(key string, hdrs []headers.Header, into []byte) []byte {
	into = append(append(into, key...), colonSpace...)

	for i := 0; i < len(hdrs)-1; i++ {
		into = append(into, hdrs[i].Value...)

		if hdrs[i].Q > 0 {
			into = append(append(into, ";q=0."...), hdrs[i].QualityString()...)
		}

		into = append(into, ',')
	}

	lastHeader := hdrs[len(hdrs)-1]
	into = append(into, lastHeader.Value...)

	if lastHeader.Q > 0 {
		into = append(append(into, ";q=0."...), lastHeader.QualityString()...)
	}

	return into
}

// isKeepAlive decides whether connection is keep-alive or not
func isKeepAlive(request *types.Request) bool {
	if request.Proto == proto.HTTP09 {
		return false
	}

	keepAlive, found := request.Headers["connection"]
	if found {
		return keepAlive[0].Value == "keep-alive"
	}

	// because HTTP/1.0 by default is not keep-alive. And if no Connection
	// is specified, it is absolutely not keep-alive
	return request.Proto == proto.HTTP11
}

// mergeHeaders simply overrides a with values from b
func mergeHeaders(a, b headers.Headers) headers.Headers {
	into := make(headers.Headers, len(a)+len(b))

	for k, v := range a {
		into[k] = v
	}

	for k, v := range b {
		into[k] = v
	}

	return into
}
