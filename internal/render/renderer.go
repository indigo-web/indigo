package render

import (
	"errors"
	"io"
	"math"
	"os"
	"strconv"
	"strings"

	"github.com/fakefloordiv/indigo/http"
	methods "github.com/fakefloordiv/indigo/http/method"
	"github.com/fakefloordiv/indigo/http/proto"
	"github.com/fakefloordiv/indigo/http/status"
	"github.com/fakefloordiv/indigo/internal/httpchars"
)

var (
	contentLength = []byte("Content-Length: ")

	errConnWrite = errors.New("error occurred while communicating with conn")
)

type (
	defaultHeader struct {
		value string
		seen  bool
	}

	headersMap = map[string]*defaultHeader
)

// Renderer is a session responses renderer. Its purpose is only to know
// something about client, and knowing them, render correct response
// for example, we SHOULD not use content-codings for HTTP/1.0 clients,
// and MUST NOT use them for HTTP/0.9 clients
// Also in case of file is being sent, it collects some meta about it,
// and compresses using available on both server and client encoders
type Renderer struct {
	buff       []byte
	buffOffset int
	keepAlive  bool

	defaultHeaders headersMap
	fileBuff       []byte
}

func NewRenderer(buff, fileBuff []byte, defaultHeaders map[string][]string) *Renderer {
	return &Renderer{
		buff:           buff,
		defaultHeaders: parseDefaultHeaders(defaultHeaders),
		fileBuff:       fileBuff,
	}
}

func (r *Renderer) Response(
	request *http.Request, response http.Response, writer http.ResponseWriter,
) (err error) {
	switch r.keepAlive {
	case false:
		// in case this value is false, this can mean only 1 thing - it is not initialized
		// and once it is initialized, it is supposed to be true, otherwise connection will
		// be closed after this call anyway
		r.keepAlive = isKeepAlive(request)
		if !r.keepAlive {
			err = status.ErrCloseConnection
		}

		r.buff = append(append(r.buff, proto.ToBytes(request.Proto)...), httpchars.SP...)
		r.buffOffset = len(r.buff)
	case true:
		// in case request Connection header is set to close, this response must be the last
		// one, after which one connection will be closed. It's better to close it silently
		if request.Headers.Value("connection") == "close" {
			err = status.ErrCloseConnection
		}
	}

	buff := r.buff[:r.buffOffset]
	codeStatus := status.CodeStatus(response.Code)

	if response.Status == "" && codeStatus != "" {
		buff = append(buff, codeStatus...)
	} else {
		// in case we have a custom response status text or unknown code, fallback to an old way
		buff = append(strconv.AppendInt(buff, int64(response.Code), 10), httpchars.SP...)
		buff = append(append(buff, status.Text(response.Code)...), httpchars.CRLF...)
	}

	responseHeaders := response.Headers()

	for i := 0; i < len(responseHeaders)/2; i += 2 {
		buff = renderHeaderInto(buff, responseHeaders[i], responseHeaders[i+1])

		if defaultHdr, found := r.defaultHeaders[responseHeaders[i]]; found {
			defaultHdr.seen = true
		}
	}

	for key, value := range r.defaultHeaders {
		if !value.seen {
			buff = renderHeaderInto(buff, key, value.value)
			value.seen = false
		}
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
	r.buff = append(buff, httpchars.CRLF...)

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
	method methods.Method, writer http.ResponseWriter, response http.Response,
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

	if err = writer(append(r.buff, httpchars.CRLF...)); err != nil {
		return errConnWrite
	}

	if method == methods.HEAD {
		// once again, HEAD requests MUST NOT contain response bodies. They are just like
		// GET request, but without response entities
		return nil
	}

	if r.fileBuff == nil {
		// write by blocks 64kb each
		r.fileBuff = make([]byte, math.MaxUint16)
	}

	for {
		n, err := file.Read(r.fileBuff)
		switch err {
		case nil:
		case io.EOF:
			return nil
		default:
			return err
		}

		if err = writer(r.fileBuff[:n]); err != nil {
			return errConnWrite
		}
	}
}

func renderContentLength(value int64, buff []byte) []byte {
	buff = append(buff, contentLength...)

	return append(strconv.AppendInt(buff, value, 10), httpchars.CRLF...)
}

func renderHeaderInto(buff []byte, key, value string) []byte {
	buff = append(buff, key...)
	buff = append(buff, httpchars.COLONSP...)
	buff = append(buff, value...)

	return append(buff, httpchars.CRLF...)
}

// isKeepAlive decides whether connection is keep-alive or not
func isKeepAlive(request *http.Request) bool {
	if request.Proto == proto.HTTP09 {
		return false
	}

	if connection := request.Headers.Value("connection"); connection != "" {
		return connection == "keep-alive"
	}

	// because HTTP/1.0 by default is not keep-alive. And if no Connection
	// is specified, it is absolutely not keep-alive
	return request.Proto == proto.HTTP11
}

func parseDefaultHeaders(hdrs map[string][]string) headersMap {
	m := make(headersMap, len(hdrs))

	for key, values := range hdrs {
		m[key] = &defaultHeader{
			value: strings.Join(values, ","),
		}
	}

	return m
}
