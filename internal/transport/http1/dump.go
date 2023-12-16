package http1

import (
	"github.com/indigo-web/indigo/http"
	"github.com/indigo-web/indigo/http/method"
	"github.com/indigo-web/indigo/http/proto"
	"github.com/indigo-web/indigo/http/status"
	"github.com/indigo-web/indigo/internal/httpchars"
	"github.com/indigo-web/indigo/internal/response"
	"github.com/indigo-web/indigo/internal/transport"
	"github.com/indigo-web/indigo/internal/transport/http1/internal/defaultheaders"
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

type Dumper struct {
	buff                  []byte
	fileBuff              []byte
	defaultHeaders        defaultheaders.DefaultHeaders
	defaultHeadersReserve defaultheaders.DefaultHeaders
	buffOffset            int
}

func NewDumper(buff, fileBuff []byte, defaultHeaders map[string][]string) *Dumper {
	parsedDefaultHeaders := parseDefaultHeaders(defaultHeaders)

	return &Dumper{
		buff:                  buff,
		fileBuff:              fileBuff,
		defaultHeadersReserve: ft.Map(ft.Nop[string], parsedDefaultHeaders), // copy the slice
		defaultHeaders:        parsedDefaultHeaders,
	}
}

// PreDump dumps the response into the buffer without actually sending it. Usually used
// for informational responses
func (d *Dumper) PreDump(protocol proto.Proto, response *http.Response) {
	d.renderProtocol(protocol)
	d.renderHeaders(response.Reveal())
	d.crlf()
}

// Dump dumps the response, keeping in mind difference between 0.9, 1.0 and 1.1 HTTP versions
func (d *Dumper) Dump(
	protocol proto.Proto, request *http.Request, response *http.Response, writer transport.Writer,
) (err error) {
	defer d.clear()

	d.renderProtocol(protocol)
	fields := response.Reveal()

	if fields.Attachment.Content() != nil {
		return d.sendAttachment(request, response, writer)
	}

	d.renderHeaders(fields)
	d.renderContentLength(int64(len(fields.Body)))
	d.crlf()

	if request.Method != method.HEAD {
		// HEAD request responses must be similar to GET request responses, except
		// forced lack of body, even if Content-Length is specified
		d.buff = append(d.buff, fields.Body...)
	}

	err = writer.Write(d.buff)

	if !isKeepAlive(protocol, request) && request.Upgrade == proto.Unknown {
		err = status.ErrCloseConnection
	}

	return err
}

func (d *Dumper) renderHeaders(fields response.Fields) {
	codeStatus := status.CodeStatus(fields.Code)

	if fields.Status == "" && codeStatus != "" {
		d.buff = append(d.buff, codeStatus...)
	} else {
		// in case we have a custom response status text or unknown code, fallback to an old way
		d.buff = append(strconv.AppendInt(d.buff, int64(fields.Code), 10), httpchars.SP...)
		d.buff = append(append(d.buff, status.Text(fields.Code)...), httpchars.CR, httpchars.LF)
	}

	responseHeaders := fields.Headers

	for i := 0; i < len(responseHeaders); i += 2 {
		d.renderHeader(responseHeaders[i], responseHeaders[i+1])
		d.defaultHeaders.EraseEntry(responseHeaders[i])
	}

	for i := 0; i < len(d.defaultHeaders); i += 2 {
		if len(d.defaultHeaders[i]) == 0 {
			continue
		}

		d.renderHeader(d.defaultHeaders[i], d.defaultHeaders[i+1])
	}

	// Content-Type is compulsory. Transfer-Encoding is not
	d.renderContentType(fields.ContentType)
	if len(fields.TransferEncoding) > 0 {
		d.renderTransferEncoding(fields.TransferEncoding)
	}
}

// sendAttachment simply encapsulates all the logic related to rendering arbitrary
// io.Reader implementations
func (d *Dumper) sendAttachment(
	request *http.Request, response *http.Response, writer transport.Writer,
) (err error) {
	fields := response.Reveal()
	size := fields.Attachment.Size()

	if size > 0 {
		d.renderHeaders(fields)
		d.renderContentLength(int64(size))
	} else {
		d.renderHeaders(response.TransferEncoding("chunked").Reveal())
	}

	// now we have to send the body via plain text or chunked transfer encoding.
	// I'm proposing to make an exception for chunked transfer encoding with a
	// separate method that'll handle with it by its own. Maybe, even for plain-text

	d.crlf()

	if err = writer.Write(d.buff); err != nil {
		return status.ErrCloseConnection
	}

	if request.Method == method.HEAD {
		// HEAD requests MUST NOT contain response bodies. They are just like
		// GET request, but without response entities
		return nil
	}

	if len(d.fileBuff) == 0 {
		// write by blocks 64kb each. Not really efficient, but in close future
		// file distributors will be implemented, so files uploading capabilities
		// will be extended
		const fileBuffSize = 128 /* kilobytes */ * 1024 /* bytes */
		d.fileBuff = make([]byte, fileBuffSize)
	}

	if fields.Attachment.Size() > 0 {
		err = d.writePlainBody(fields.Attachment.Content(), writer)
	} else {
		err = d.writeChunkedBody(fields.Attachment.Content(), writer)
	}

	fields.Attachment.Close()

	return err
}

func (d *Dumper) writePlainBody(r io.Reader, writer transport.Writer) error {
	// TODO: implement checking whether r implements io.ReaderAt interfacd. In case it does
	//       body may be transferred more efficiently. This requires implementing io.Writer
	//       *http.ResponseWriter

	for {
		n, err := r.Read(d.fileBuff)
		switch err {
		case nil:
		case io.EOF:
			return nil
		default:
			return status.ErrCloseConnection
		}

		if err = writer.Write(d.fileBuff[:n]); err != nil {
			return status.ErrCloseConnection
		}
	}
}

func (d *Dumper) writeChunkedBody(r io.Reader, writer transport.Writer) error {
	const (
		hexValueOffset = 8
		crlfSize       = 1 /* CR */ + 1 /* LF */
		buffOffset     = hexValueOffset + crlfSize
	)

	for {
		n, err := r.Read(d.fileBuff[buffOffset : len(d.fileBuff)-crlfSize])

		if n > 0 {
			// first rewrite begin of the fileBuff to contain our hexdecimal value
			buff := strconv.AppendUint(d.fileBuff[:0], uint64(n), 16)
			// now we can determine the length of the hexdecimal value and make an
			// offset for it
			blankSpace := hexValueOffset - len(buff)
			copy(d.fileBuff[blankSpace:], buff)
			copy(d.fileBuff[hexValueOffset:], httpchars.CRLF)
			copy(d.fileBuff[buffOffset+n:], httpchars.CRLF)

			if err := writer.Write(d.fileBuff[blankSpace : buffOffset+n+crlfSize]); err != nil {
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
func (d *Dumper) renderHeader(key, value string) {
	d.buff = append(d.buff, key...)
	d.buff = append(d.buff, httpchars.COLONSP...)
	d.buff = append(d.buff, value...)
	d.crlf()
}

func (d *Dumper) renderContentLength(value int64) {
	d.buff = strconv.AppendInt(append(d.buff, contentLength...), value, 10)
	d.crlf()
}

func (d *Dumper) renderContentType(value string) {
	d.buff = append(d.buff, contentType...)
	d.buff = append(d.buff, value...)
	d.crlf()
}

func (d *Dumper) renderTransferEncoding(value string) {
	d.buff = append(d.buff, transferEncoding...)
	d.buff = append(d.buff, value...)
	d.crlf()
}

func (d *Dumper) renderProtocol(protocol proto.Proto) {
	d.buff = append(d.buff, proto.ToBytes(protocol)...)
}

func (d *Dumper) crlf() {
	d.buff = append(d.buff, httpchars.CRLF...)
}

func (d *Dumper) clear() {
	d.buff = d.buff[:0]
	d.defaultHeadersReserve.Copy(d.defaultHeaders)
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

func parseDefaultHeaders(hdrs map[string][]string) defaultheaders.DefaultHeaders {
	parsed := make([]string, 0, len(hdrs))

	for key, values := range hdrs {
		value := strings.Join(values, ",")
		parsed = append(parsed, key, value)
	}

	return parsed
}
