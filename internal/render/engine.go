package render

import (
	"errors"
	"github.com/indigo-web/indigo/http/status"
	"github.com/indigo-web/indigo/internal/render/types"
	"io"
	"math"
	"os"
	"strconv"
	"strings"

	"github.com/indigo-web/indigo/http"
	methods "github.com/indigo-web/indigo/http/method"
	"github.com/indigo-web/indigo/http/proto"
	"github.com/indigo-web/indigo/internal/httpchars"
)

var (
	contentLength = []byte("Content-Length: ")

	errConnWrite    = errors.New("error occurred while communicating with conn")
	errFileNotFound = errors.New("desired file not found")
)

type Engine interface {
	PreWrite(*http.Request, http.Response)
	Write(*http.Request, http.Response, http.ResponseWriter) error
}

// engine is a renderer engine for HTTP responses. The point of it is to render response
// into the buffer that'll be written to the socket later. Engine also owns connection
// object, that actually breaks SRP, but compulsory as we must provide enough flexibility
// to make possible files distribution be more efficient. At the moment, this feature isn't
// implemented yet, but will be soon
type engine struct {
	buff, fileBuff []byte
	buffOffset     int
	defaultHeaders types.HeadersMap

	// TODO: add files distribution mechanism (and edit docstring)
}

func NewEngine(buff, fileBuff []byte, defaultHeaders map[string][]string) Engine {
	return &engine{
		buff:           buff,
		fileBuff:       fileBuff,
		defaultHeaders: parseDefaultHeaders(defaultHeaders),
	}
}

func (e *engine) PreWrite(request *http.Request, response http.Response) {
	if e.buffOffset == 0 {
		e.renderProtocol(request.Proto)
	}

	e.renderHeaders(response)
	e.crlf()
}

// Render the response, respectively to the protocol
func (e *engine) Write(
	request *http.Request, response http.Response, writer http.ResponseWriter,
) (err error) {
	defer e.clear()

	if e.buffOffset == 0 {
		e.renderProtocol(request.Proto)
	}

	if path, errhandler := response.File(); len(path) > 0 {
		switch err := e.renderFile(request, response, writer); err {
		case errFileNotFound:
			return e.Write(request, errhandler(status.ErrNotFound), writer)
		default:
			// nil will also be returned here
			return err
		}
	}

	e.renderHeaders(response)
	e.renderContentLength(int64(len(response.Body)))
	e.crlf()

	if request.Method != methods.HEAD {
		// HEAD request responses must be similar to GET request responses, except
		// forced lack of body, even if Content-Length is specified
		e.buff = append(e.buff, response.Body...)
	}

	err = writer(e.buff)

	if err == nil && !isKeepAlive(request) {
		err = errConnWrite
	}

	return err
}

func (e *engine) renderHeaders(response http.Response) {
	codeStatus := status.CodeStatus(response.Code)

	if response.Status == "" && codeStatus != "" {
		e.buff = append(e.buff, codeStatus...)
	} else {
		// in case we have a custom response status text or unknown code, fallback to an old way
		e.buff = append(strconv.AppendInt(e.buff, int64(response.Code), 10), httpchars.SP...)
		e.buff = append(append(e.buff, status.Text(response.Code)...), httpchars.CRLF...)
	}

	responseHeaders := response.Headers()

	for i := 0; i < len(responseHeaders)/2; i += 2 {
		e.renderHeader(responseHeaders[i], responseHeaders[i+1])

		if defaultHdr, found := e.defaultHeaders[responseHeaders[i]]; found {
			defaultHdr.Seen = true
		}
	}

	for key, value := range e.defaultHeaders {
		if !value.Seen {
			e.renderHeader(key, value.Value)
			value.Seen = false
		}
	}
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
func (e *engine) renderFile(
	request *http.Request, response http.Response, writer http.ResponseWriter,
) error {
	filename, _ := response.File()
	file, err := os.OpenFile(filename, os.O_RDONLY, 0)
	if err != nil {
		return errFileNotFound
	}

	stat, err := file.Stat()
	if err != nil {
		return err
	}

	e.renderHeaders(response)
	e.renderContentLength(stat.Size())
	e.crlf()

	if err = writer(e.buff); err != nil {
		return errConnWrite
	}

	if request.Method == methods.HEAD {
		// HEAD requests MUST NOT contain response bodies. They are just like
		// GET request, but without response entities
		return nil
	}

	if e.fileBuff == nil {
		// write by blocks 64kb each. Not really efficient, but in close future
		// file distributors will be implemented, so files uploading capabilities
		// will be extended
		e.fileBuff = make([]byte, math.MaxUint16)
	}

	for {
		n, err := file.Read(e.fileBuff)
		switch err {
		case nil:
		case io.EOF:
			return nil
		default:
			return errConnWrite
		}

		if err = writer(e.fileBuff[:n]); err != nil {
			return errConnWrite
		}
	}
}

// renderHeaderInto the buffer. Appends CRLF in the end
func (e *engine) renderHeader(key, value string) {
	e.buff = append(e.buff, key...)
	e.buff = append(e.buff, httpchars.COLONSP...)
	e.buff = append(e.buff, value...)
	e.crlf()
}

func (e *engine) renderContentLength(value int64) {
	e.buff = strconv.AppendInt(append(e.buff, contentLength...), value, 10)
	e.crlf()
}

func (e *engine) renderProtocol(protocol proto.Proto) {
	e.buff = append(e.buff, proto.ToBytes(protocol)...)
	e.buffOffset = len(e.buff)
}

func (e *engine) crlf() {
	e.buff = append(e.buff, httpchars.CRLF...)
}

func (e *engine) clear() {
	e.buff = e.buff[:e.buffOffset]
}

func isKeepAlive(req *http.Request) bool {
	switch req.Proto {
	case proto.HTTP09, proto.HTTP10:
		// actually, HTTP/0.9 doesn't even have a Connection: keep-alive header,
		// but who knows - let it be
		return strings.EqualFold(req.Headers.Value("connection"), "keep-alive")
	case proto.HTTP11:
		// in case of HTTP/1.1, keep-alive may be only disabled
		return !strings.EqualFold(req.Headers.Value("connection"), "close")
	case proto.HTTP2:
		// TODO: are there cases when HTTP/2 connection may not be keep-alived?
		return true
	default:
		// don't know what this is, but act like everything is okay
		return true
	}
}
func parseDefaultHeaders(hdrs map[string][]string) types.HeadersMap {
	m := make(types.HeadersMap, len(hdrs))

	for key, values := range hdrs {
		m[key] = &types.DefaultHeader{
			Value: strings.Join(values, ","),
		}
	}

	return m
}
