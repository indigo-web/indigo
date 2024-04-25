package http1

import (
	"github.com/indigo-web/indigo/http"
	"github.com/indigo-web/indigo/http/cookie"
	"github.com/indigo-web/indigo/http/headers"
	"github.com/indigo-web/indigo/http/method"
	"github.com/indigo-web/indigo/http/proto"
	"github.com/indigo-web/indigo/http/status"
	"github.com/indigo-web/indigo/internal/httpchars"
	"github.com/indigo-web/indigo/internal/response"
	"github.com/indigo-web/utils/strcomp"
	"github.com/indigo-web/utils/uf"
	"io"
	"log"
	"strconv"
	"time"
)

const (
	contentType      = "Content-Type: "
	transferEncoding = "Transfer-Encoding: "
	contentLength    = "Content-Length: "
	setCookie        = "Set-Cookie: "
)

// minimalFileBuffSize defines the minimal size of the file buffer. In case it's less
// it'll be set to this value and debug log will be printed
const minimalFileBuffSize = 16

var (
	chunkedFinalizer = []byte("0\r\n\r\n")
	gmt              = time.FixedZone("GMT", 0)
)

type Writer interface {
	Write([]byte) error
}

type Serializer struct {
	request *http.Request
	writer  Writer
	buff    []byte
	// fileBuff isn't allocated until needed in order to save memory in cases,
	// where no files are being sent
	fileBuff       []byte
	fileBuffSize   int
	defaultHeaders defaultHeaders
}

func NewSerializer(
	buff []byte,
	fileBuffSize int,
	defHdrs map[string]string,
	request *http.Request,
	writer Writer,
) *Serializer {
	if fileBuffSize < minimalFileBuffSize {
		log.Printf("misconfiguration: file buffer size (Config.HTTP.FileBuffSize) is %d, "+
			"which is below minimal (%d). The value is forcefully set to %d\n",
			fileBuffSize, minimalFileBuffSize, minimalFileBuffSize,
		)

		fileBuffSize = minimalFileBuffSize
	}

	return &Serializer{
		request:        request,
		writer:         writer,
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

// Write writes the response, keeping in mind difference between 1.0 and 1.1 HTTP versions
func (d *Serializer) Write(
	protocol proto.Proto, response *http.Response,
) (err error) {
	defer d.clear()

	d.renderProtocol(protocol)
	fields := response.Reveal()
	d.renderResponseLine(fields)

	if fields.Attachment.Content() != nil {
		return d.sendAttachment(d.request, response, d.writer)
	}

	d.renderHeaders(fields)

	for _, c := range fields.Cookies {
		d.renderCookie(c)
	}

	d.renderContentLength(int64(len(fields.Body)))
	d.crlf()

	if d.request.Method != method.HEAD {
		// HEAD request responses must be similar to GET request responses, except
		// forced lack of body, even if Content-Length is specified
		d.buff = append(d.buff, fields.Body...)
	}

	err = d.writer.Write(d.buff)

	if !isKeepAlive(protocol, d.request) && d.request.Upgrade == proto.Unknown {
		err = status.ErrCloseConnection
	}

	return err
}

func (d *Serializer) renderResponseLine(fields *response.Fields) {
	statusLine := status.Line(fields.Code)

	if fields.Status == "" && statusLine != "" {
		d.buff = append(d.buff, statusLine...)
		return
	}

	// in case we have a custom response status text or unknown code, fallback to an old way
	d.buff = strconv.AppendInt(d.buff, int64(fields.Code), 10)
	d.sp()
	d.buff = append(d.buff, status.Text(fields.Code)...)
	d.crlf()
}

func (d *Serializer) renderHeaders(fields *response.Fields) {
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
	request *http.Request, response *http.Response, writer Writer,
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

func (d *Serializer) writePlainBody(r io.Reader, writer Writer) error {
	// TODO: implement checking whether r implements io.ReaderAt interface. In case it does
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

func (d *Serializer) writeChunkedBody(r io.Reader, writer Writer) error {
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
func (d *Serializer) renderHeader(header headers.Header) {
	d.buff = append(d.buff, header.Key...)
	d.colonsp()
	d.buff = append(d.buff, header.Value...)
	d.crlf()
}

func (d *Serializer) renderCookie(c cookie.Cookie) {
	d.buff = append(d.buff, setCookie...)
	d.buff = append(d.buff, c.Name...)
	d.buff = append(d.buff, '=')
	d.buff = append(d.buff, c.Value...)
	d.buff = append(d.buff, ';', ' ')

	if len(c.Path) > 0 {
		d.buff = append(d.buff, "Path="...)
		d.buff = append(d.buff, c.Path...)
		d.buff = append(d.buff, ';', ' ')
	}

	if len(c.Domain) > 0 {
		d.buff = append(d.buff, "Domain="...)
		d.buff = append(d.buff, c.Domain...)
		d.buff = append(d.buff, ';', ' ')
	}

	if !c.Expires.IsZero() {
		d.buff = append(d.buff, "Expires="...)
		// TODO: this will probably be slow. Can be optimized via rendering it manually
		//  directly into the d.buff
		d.buff = append(d.buff, c.Expires.In(gmt).Format(time.RFC1123)...)
		d.buff = append(d.buff, ';', ' ')
	}

	if c.MaxAge != 0 {
		maxage := "0"
		if c.MaxAge > 0 {
			maxage = strconv.Itoa(c.MaxAge)
		}

		d.buff = append(d.buff, "MaxAge="...)
		d.buff = append(d.buff, maxage...)
		d.buff = append(d.buff, ';', ' ')
	}

	if len(c.SameSite) > 0 {
		d.buff = append(d.buff, "SameSite="...)
		d.buff = append(d.buff, c.SameSite...)
		d.buff = append(d.buff, ';', ' ')
	}

	if c.Secure {
		d.buff = append(d.buff, "Secure; "...)
	}

	if c.HttpOnly {
		d.buff = append(d.buff, "HttpOnly; "...)
	}

	// strip last 2 bytes, which are always a semicolon and a space
	d.buff = d.buff[:len(d.buff)-2]
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
	d.buff = append(d.buff, protocol.String()...)
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
	case proto.HTTP10:
		return strcomp.EqualFold(req.Connection, "keep-alive")
	case proto.HTTP11:
		// in case of HTTP/1.1, keep-alive may be only disabled
		return !strcomp.EqualFold(req.Connection, "close")
	default:
		// as the protocol is unknown and the code was probably caused by some sort
		// of bug, consider closing it
		return false
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
