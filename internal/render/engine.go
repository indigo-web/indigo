package render

import (
	"github.com/indigo-web/indigo/http"
	"github.com/indigo-web/indigo/http/method"
	"github.com/indigo-web/indigo/http/proto"
	"github.com/indigo-web/indigo/http/status"
	"github.com/indigo-web/indigo/internal/httpchars"
	"github.com/indigo-web/indigo/internal/render/types"
	"github.com/indigo-web/indigo/internal/response"
	"github.com/indigo-web/utils/ft"
	"github.com/indigo-web/utils/strcomp"
	"io"
	"strconv"
	"strings"
)

var (
	contentLength    = []byte("Content-Length: ")
	contentType      = []byte("Content-Type: ")
	transferEncoding = []byte("Transfer-Encoding: ")
	emptyChunkedPart = []byte("0\r\n\r\n")
)

type ClientWriter interface {
	Write([]byte) error
}

type Engine interface {
	PreWrite(proto.Proto, *http.Response)
	Write(proto.Proto, *http.Request, *http.Response, ClientWriter) error
}

// engine is a renderer engine for HTTP responses. The point of it is to render response
// into the buffer that'll be written to the socket later. Engine also owns connection
// object, that actually breaks SRP, but compulsory as we must provide enough flexibility
// to make possible files distribution be more efficient. At the moment, this feature isn't
// implemented yet, but will be soon
type engine struct {
	buff                  []byte
	fileBuff              []byte
	defaultHeaders        types.DefaultHeaders
	defaultHeadersReserve types.DefaultHeaders
	buffOffset            int
}

func NewEngine(buff, fileBuff []byte, defaultHeaders map[string][]string) Engine {
	return newEngine(buff, fileBuff, defaultHeaders)
}

func newEngine(buff, fileBuff []byte, defaultHeaders map[string][]string) *engine {
	parsedDefaultHeaders := parseDefaultHeaders(defaultHeaders)

	return &engine{
		buff:                  buff,
		fileBuff:              fileBuff,
		defaultHeadersReserve: ft.Map(ft.Nop[string], parsedDefaultHeaders),
		defaultHeaders:        parsedDefaultHeaders,
	}
}

// PreWrite writes the response into the buffer without actually sending it. Usually used
// for informational responses
func (e *engine) PreWrite(protocol proto.Proto, response *http.Response) {
	e.renderProtocol(protocol)
	e.renderHeaders(response.Reveal())
	e.crlf()
}

// Render the response, respectively to the protocol
func (e *engine) Write(
	protocol proto.Proto, request *http.Request, response *http.Response, writer ClientWriter,
) (err error) {
	defer e.clear()

	e.renderProtocol(protocol)
	fields := response.Reveal()

	if fields.Attachment.Content() != nil {
		return e.sendAttachment(request, response, writer)
	}

	e.renderHeaders(fields)
	e.renderContentLength(int64(len(fields.Body)))
	e.crlf()

	if request.Method != method.HEAD {
		// HEAD request responses must be similar to GET request responses, except
		// forced lack of body, even if Content-Length is specified
		e.buff = append(e.buff, fields.Body...)
	}

	err = writer.Write(e.buff)

	if !isKeepAlive(protocol, request) && request.Upgrade == proto.Unknown {
		err = status.ErrCloseConnection
	}

	return err
}

func (e *engine) renderHeaders(fields response.Fields) {
	codeStatus := status.CodeStatus(fields.Code)

	if fields.Status == "" && codeStatus != "" {
		e.buff = append(e.buff, codeStatus...)
	} else {
		// in case we have a custom response status text or unknown code, fallback to an old way
		e.buff = append(strconv.AppendInt(e.buff, int64(fields.Code), 10), httpchars.SP...)
		e.buff = append(append(e.buff, status.Text(fields.Code)...), httpchars.CR, httpchars.LF)
	}

	responseHeaders := fields.Headers

	for i := 0; i < len(responseHeaders); i += 2 {
		e.renderHeader(responseHeaders[i], responseHeaders[i+1])
		e.defaultHeaders.EraseEntry(responseHeaders[i])
	}

	for i := 0; i < len(e.defaultHeaders); i += 2 {
		if len(e.defaultHeaders[i]) == 0 {
			continue
		}

		e.renderHeader(e.defaultHeaders[i], e.defaultHeaders[i+1])
	}

	// Content-Type is compulsory. Transfer-Encoding is not
	e.renderContentType(fields.ContentType)
	if len(fields.TransferEncoding) > 0 {
		e.renderTransferEncoding(fields.TransferEncoding)
	}
}

// sendAttachment simply encapsulates all the logic related to rendering arbitrary
// io.Reader implementations
func (e *engine) sendAttachment(
	request *http.Request, response *http.Response, writer ClientWriter,
) (err error) {
	fields := response.Reveal()
	size := fields.Attachment.Size()

	if size > 0 {
		e.renderHeaders(fields)
		e.renderContentLength(int64(size))
	} else {
		e.renderHeaders(response.TransferEncoding("chunked").Reveal())
	}

	// now we have to send the body via plain text or chunked transfer encoding.
	// I'm proposing to make an exception for chunked transfer encoding with a
	// separate method that'll handle with it by its own. Maybe, even for plain-text

	e.crlf()

	if err = writer.Write(e.buff); err != nil {
		return status.ErrCloseConnection
	}

	if request.Method == method.HEAD {
		// HEAD requests MUST NOT contain response bodies. They are just like
		// GET request, but without response entities
		return nil
	}

	if len(e.fileBuff) == 0 {
		// write by blocks 64kb each. Not really efficient, but in close future
		// file distributors will be implemented, so files uploading capabilities
		// will be extended
		const fileBuffSize = 128 /* kilobytes */ * 1024 /* bytes */
		e.fileBuff = make([]byte, fileBuffSize)
	}

	if fields.Attachment.Size() > 0 {
		err = e.writePlainBody(fields.Attachment.Content(), writer)
	} else {
		err = e.writeChunkedBody(fields.Attachment.Content(), writer)
	}

	fields.Attachment.Close()

	return err
}

func (e *engine) writePlainBody(r io.Reader, writer ClientWriter) error {
	// TODO: implement checking whether r implements io.ReaderAt interface. In case it does
	//       body may be transferred more efficiently. This requires implementing io.Writer
	//       *http.ResponseWriter

	for {
		n, err := r.Read(e.fileBuff)
		switch err {
		case nil:
		case io.EOF:
			return nil
		default:
			return status.ErrCloseConnection
		}

		if err = writer.Write(e.fileBuff[:n]); err != nil {
			return status.ErrCloseConnection
		}
	}
}

func (e *engine) writeChunkedBody(r io.Reader, writer ClientWriter) error {
	const (
		hexValueOffset = 8
		crlfSize       = 1 /* CR */ + 1 /* LF */
		buffOffset     = hexValueOffset + crlfSize
	)

	for {
		n, err := r.Read(e.fileBuff[buffOffset : len(e.fileBuff)-crlfSize])

		if n > 0 {
			// first rewrite begin of the fileBuff to contain our hexdecimal value
			buff := strconv.AppendUint(e.fileBuff[:0], uint64(n), 16)
			// now we can determine the length of the hexdecimal value and make an
			// offset for it
			blankSpace := hexValueOffset - len(buff)
			copy(e.fileBuff[blankSpace:], buff)
			copy(e.fileBuff[hexValueOffset:], httpchars.CRLF)
			copy(e.fileBuff[buffOffset+n:], httpchars.CRLF)

			if err := writer.Write(e.fileBuff[blankSpace : buffOffset+n+crlfSize]); err != nil {
				return status.ErrCloseConnection
			}
		}

		switch err {
		case nil:
		case io.EOF:
			return writer.Write(emptyChunkedPart)
		default:
			return status.ErrCloseConnection
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

func (e *engine) renderContentType(value string) {
	e.buff = append(e.buff, contentType...)
	e.buff = append(e.buff, value...)
	e.crlf()
}

func (e *engine) renderTransferEncoding(value string) {
	e.buff = append(e.buff, transferEncoding...)
	e.buff = append(e.buff, value...)
	e.crlf()
}

func (e *engine) renderProtocol(protocol proto.Proto) {
	e.buff = append(e.buff, proto.ToBytes(protocol)...)
}

func (e *engine) crlf() {
	e.buff = append(e.buff, httpchars.CRLF...)
}

func (e *engine) clear() {
	e.buff = e.buff[:0]
	e.defaultHeadersReserve.Copy(e.defaultHeaders)
}

func isKeepAlive(protocol proto.Proto, req *http.Request) bool {
	switch protocol {
	case proto.HTTP09, proto.HTTP10:
		// actually, HTTP/0.9 doesn't even have a Connection: keep-alive header,
		// but who knows - let it be
		return strcomp.EqualFold(req.Headers.Value("connection"), "keep-alive")
	case proto.HTTP11:
		// in case of HTTP/1.1, keep-alive may be only disabled
		return !strcomp.EqualFold(req.Headers.Value("connection"), "close")
	case proto.HTTP2:
		// TODO: are there cases when HTTP/2 connection may not be keep-alived?
		return true
	default:
		// don't know what this is, but act like everything is okay
		return true
	}
}

func parseDefaultHeaders(hdrs map[string][]string) types.DefaultHeaders {
	parsed := make([]string, 0, len(hdrs))

	for key, values := range hdrs {
		value := strings.Join(values, ",")
		parsed = append(parsed, key, value)
	}

	return parsed
}
