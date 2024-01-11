package http1

import (
	"github.com/indigo-web/indigo/http"
	"github.com/indigo-web/indigo/http/method"
	"github.com/indigo-web/indigo/http/proto"
	"github.com/indigo-web/indigo/http/status"
	"github.com/indigo-web/indigo/internal/httpchars"
	"github.com/indigo-web/indigo/internal/response"
	"github.com/indigo-web/indigo/internal/transport"
	"github.com/indigo-web/utils/strcomp"
	"github.com/indigo-web/utils/uf"
	"io"
	"log"
	"strconv"
)

const (
	contentType      = "Content-Type: "
	transferEncoding = "Transfer-Encoding: "
	contentLength    = "Content-Length: "
)

// minimalFileBuffSize defines the minimal size of the file buffer. In case it's less
// it'll be set to this value and debug log will be printed
const minimalFileBuffSize = 16

var chunkedFinalizer = []byte("0\r\n\r\n")

type Serializer struct {
	buff []byte
	// fileBuff isn't allocated until needed in order to save memory in cases,
	// where no files are being sent
	fileBuff       []byte
	fileBuffSize   int
	defaultHeaders defaultHeaders
}

func NewSerializer(buff []byte, fileBuffSize int, defHdrs map[string]string) *Serializer {
	if fileBuffSize < minimalFileBuffSize {
		log.Printf("misconfiguration: file buffer size (Settings.HTTP.FileBuffSize) is set to %d, "+
			"however minimal possible value is %d. Setting it hard to %d\n",
			fileBuffSize, minimalFileBuffSize, minimalFileBuffSize,
		)

		fileBuffSize = minimalFileBuffSize
	}

	return &Serializer{
		buff:           buff[:0],
		fileBuffSize:   fileBuffSize,
		defaultHeaders: processDefaultHeaders(defHdrs),
	}
}

// PreWrite writes the response into the buffer without actually sending it. Usually used
// for informational responses
func (d *Serializer) PreWrite(protocol proto.Proto, response *http.Response) {
	d.renderProtocol(protocol)
	fields := response.Reveal()
	d.renderResponseLine(fields)
	d.renderHeaders(fields)
	d.crlf()
}

// Write writes the response, keeping in mind difference between 0.9, 1.0 and 1.1 HTTP versions
func (d *Serializer) Write(
	protocol proto.Proto, request *http.Request, response *http.Response, writer transport.Writer,
) (err error) {
	defer d.clear()

	d.renderProtocol(protocol)
	fields := response.Reveal()
	d.renderResponseLine(fields)

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

func (d *Serializer) renderResponseLine(fields response.Fields) {
	codeStatus := status.CodeStatus(fields.Code)

	if fields.Status == "" && codeStatus != "" {
		d.buff = append(d.buff, codeStatus...)
		return
	}

	// in case we have a custom response status text or unknown code, fallback to an old way
	d.buff = strconv.AppendInt(d.buff, int64(fields.Code), 10)
	d.sp()
	d.buff = append(d.buff, status.Text(fields.Code)...)
	d.crlf()
}

func (d *Serializer) renderHeaders(fields response.Fields) {
	responseHeaders := fields.Headers

	for _, header := range responseHeaders {
		d.renderHeader(header)
		d.defaultHeaders.Exclude(header.Key)
	}

	for _, header := range d.defaultHeaders {
		if header.Excluded {
			continue
		}

		d.buff = append(d.buff, header.Full...)
	}

	// Content-Type is compulsory. Transfer-Encoding is not
	d.renderKnownHeader(contentType, fields.ContentType)
	if len(fields.TransferEncoding) > 0 {
		d.renderKnownHeader(transferEncoding, fields.TransferEncoding)
	}
}

// sendAttachment simply encapsulates all the logic related to rendering arbitrary
// io.Reader implementations
func (d *Serializer) sendAttachment(
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
		d.fileBuff = make([]byte, d.fileBuffSize)
	}

	if fields.Attachment.Size() > 0 {
		err = d.writePlainBody(fields.Attachment.Content(), writer)
	} else {
		err = d.writeChunkedBody(fields.Attachment.Content(), writer)
	}

	fields.Attachment.Close()

	return err
}

func (d *Serializer) writePlainBody(r io.Reader, writer transport.Writer) error {
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

func (d *Serializer) writeChunkedBody(r io.Reader, writer transport.Writer) error {
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
			return writer.Write(chunkedFinalizer)
		default:
			return status.ErrCloseConnection
		}
	}
}

// renderHeaderInto the buffer. Appends CRLF in the end
func (d *Serializer) renderHeader(header response.Header) {
	d.buff = append(d.buff, header.Key...)
	d.colonsp()
	d.buff = append(d.buff, header.Value...)
	d.crlf()
}

func (d *Serializer) renderContentLength(value int64) {
	d.buff = strconv.AppendInt(append(d.buff, contentLength...), value, 10)
	d.crlf()
}

func (d *Serializer) renderKnownHeader(key, value string) {
	d.buff = append(d.buff, key...)
	d.buff = append(d.buff, value...)
	d.crlf()
}

func (d *Serializer) renderProtocol(protocol proto.Proto) {
	d.buff = append(d.buff, proto.ToBytes(protocol)...)
}

func (d *Serializer) sp() {
	d.buff = append(d.buff, ' ')
}

func (d *Serializer) colonsp() {
	d.buff = append(d.buff, httpchars.COLONSP...)
}

func (d *Serializer) crlf() {
	d.buff = append(d.buff, httpchars.CRLF...)
}

func (d *Serializer) clear() {
	d.buff = d.buff[:0]
	d.defaultHeaders.Reset()
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

func processDefaultHeaders(hdrs map[string]string) defaultHeaders {
	processed := make(defaultHeaders, 0, len(hdrs))

	for key, value := range hdrs {
		full := renderHeader(key, value)
		processed = append(processed, defaultHeader{
			// we let the GC release all the values of the map, as here we're using only
			// the brand-new line without keeping the original string
			Key:  full[:len(key)],
			Full: full,
		})
	}

	return processed
}

func renderHeader(key, value string) string {
	return key + httpchars.COLONSP + value + uf.B2S(httpchars.CRLF)
}

type defaultHeader struct {
	Excluded bool
	Key      string
	Full     string
}

type defaultHeaders []defaultHeader

func (d defaultHeaders) Exclude(key string) {
	for i, header := range d {
		if strcomp.EqualFold(header.Key, key) {
			header.Excluded = true
			d[i] = header
			return
		}
	}
}

func (d defaultHeaders) Reset() {
	for _, header := range d {
		header.Excluded = false
	}
}
