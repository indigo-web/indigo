package http1

import (
	"fmt"
	"github.com/indigo-web/indigo/http"
	"github.com/indigo-web/indigo/http/cookie"
	"github.com/indigo-web/indigo/http/headers"
	"github.com/indigo-web/indigo/http/method"
	"github.com/indigo-web/indigo/http/proto"
	"github.com/indigo-web/indigo/http/status"
	"github.com/indigo-web/indigo/internal/response"
	"github.com/indigo-web/utils/strcomp"
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
	crlf             = "\r\n"
)

// minimalFileBuffSize defines the minimal size of the file buffer. In case it's less
// it'll be set to this value and debug log will be printed
const minimalFileBuffSize = 16

var (
	chunkedFinalizer = []byte("0\r\n\r\n")
	gmt              = time.FixedZone("GMT", 0)
)

type serializer struct {
	request *http.Request
	writer  io.Writer
	buff    []byte
	// fileBuff isn't allocated until needed in order to save memory in cases,
	// where no files are being sent
	fileBuff       []byte
	fileBuffSize   int
	defaultHeaders defaultHeaders
}

func newSerializer(
	buff []byte,
	fileBuffSize int,
	defHdrs map[string]string,
	request *http.Request,
	writer io.Writer,
) *serializer {
	if fileBuffSize < minimalFileBuffSize {
		log.Printf("misconfiguration: file buffer size (Config.HTTP.FileBuffSize) is %d, "+
			"which is below minimal (%d). The value is forcefully set to %d\n",
			fileBuffSize, minimalFileBuffSize, minimalFileBuffSize,
		)

		fileBuffSize = minimalFileBuffSize
	}

	return &serializer{
		request:        request,
		writer:         writer,
		buff:           buff[:0],
		fileBuffSize:   fileBuffSize,
		defaultHeaders: processDefaultHeaders(defHdrs),
	}
}

// PreWrite writes the response into the buffer without actually sending it. Usually used
// for informational responses
func (s *serializer) PreWrite(protocol proto.Proto, response *http.Response) {
	s.renderProtocol(protocol)
	fields := response.Reveal()
	s.renderResponseLine(fields)
	s.renderHeaders(fields)
	s.crlf()
}

// Write writes the response, keeping in mind difference between 1.0 and 1.1 HTTP versions
func (s *serializer) Write(
	protocol proto.Proto, response *http.Response,
) (err error) {
	defer s.clear()

	s.renderProtocol(protocol)
	fields := response.Reveal()
	s.renderResponseLine(fields)

	if fields.Attachment.Content() != nil {
		return s.sendAttachment(s.request, response, s.writer)
	}

	s.renderHeaders(fields)

	for _, c := range fields.Cookies {
		s.renderCookie(c)
	}

	s.renderContentLength(int64(len(fields.Body)))
	s.crlf()

	if s.request.Method != method.HEAD {
		// HEAD request responses must be similar to GET request responses, except
		// forced lack of body, even if Content-Length is specified
		s.buff = append(s.buff, fields.Body...)
	}

	_, err = s.writer.Write(s.buff)

	if !isKeepAlive(protocol, s.request) && s.request.Upgrade == proto.Unknown {
		err = status.ErrCloseConnection
	}

	return err
}

func (s *serializer) renderResponseLine(fields *response.Fields) {
	statusLine := status.Line(fields.Code)

	if fields.Status == "" && statusLine != "" {
		s.buff = append(s.buff, statusLine...)
		return
	}

	// in case we have a custom response status text or unknown code, fallback to an old way
	s.buff = strconv.AppendInt(s.buff, int64(fields.Code), 10)
	s.sp()
	s.buff = append(s.buff, status.Text(fields.Code)...)
	s.crlf()
}

func (s *serializer) renderHeaders(fields *response.Fields) {
	responseHeaders := fields.Headers

	for _, header := range responseHeaders {
		s.renderHeader(header)
		s.defaultHeaders.Exclude(header.Key)
	}

	for _, header := range s.defaultHeaders {
		if header.Excluded {
			continue
		}

		s.buff = append(s.buff, header.Full...)
	}

	// Content-Type is compulsory. Transfer-Encoding is not
	s.renderKnownHeader(contentType, fields.ContentType)
	if len(fields.TransferEncoding) > 0 {
		s.renderKnownHeader(transferEncoding, fields.TransferEncoding)
	}
}

// sendAttachment simply encapsulates all the logic related to rendering arbitrary
// io.Reader implementations
func (s *serializer) sendAttachment(
	request *http.Request, response *http.Response, writer io.Writer,
) (err error) {
	fields := response.Reveal()
	size := fields.Attachment.Size()

	if size > 0 {
		s.renderHeaders(fields)
		s.renderContentLength(int64(size))
	} else {
		s.renderHeaders(response.TransferEncoding("chunked").Reveal())
	}

	// now we have to send the body via plain text or chunked transfer encoding.
	// I'm proposing to make an exception for chunked transfer encoding with a
	// separate method that'll handle with it by its own. Maybe, even for plain-text

	s.crlf()

	if _, err = writer.Write(s.buff); err != nil {
		return status.ErrCloseConnection
	}

	if request.Method == method.HEAD {
		// HEAD requests MUST NOT contain response bodies. They are just like
		// GET request, but without response entities
		return nil
	}

	if len(s.fileBuff) == 0 {
		s.fileBuff = make([]byte, s.fileBuffSize)
	}

	if fields.Attachment.Size() > 0 {
		err = s.writePlainBody(fields.Attachment.Content(), writer)
	} else {
		err = s.writeChunkedBody(fields.Attachment.Content(), writer)
	}

	fields.Attachment.Close()

	return err
}

func (s *serializer) writePlainBody(r io.Reader, writer io.Writer) error {
	// TODO: we really could simply use the response buffer instead of a separated file buffer.

	if w, ok := r.(io.WriterTo); ok {
		_, err := w.WriteTo(writer)
		if err != nil {
			// ignore any occurred errors, cut the connection down immediately
			err = status.ErrCloseConnection
		}

		return err
	}

	for {
		n, err := r.Read(s.fileBuff)
		switch err {
		case nil:
		case io.EOF:
			return nil
		default:
			return status.ErrCloseConnection
		}

		if _, err = writer.Write(s.fileBuff[:n]); err != nil {
			return status.ErrCloseConnection
		}
	}
}

func (s *serializer) writeChunkedBody(r io.Reader, writer io.Writer) error {
	const (
		hexValueOffset = 8
		crlfSize       = 1 /* CR */ + 1 /* LF */
		buffOffset     = hexValueOffset + crlfSize
	)

	for {
		n, err := r.Read(s.fileBuff[buffOffset : len(s.fileBuff)-crlfSize])

		if n > 0 {
			// first rewrite begin of the fileBuff to contain our hexdecimal value
			buff := strconv.AppendUint(s.fileBuff[:0], uint64(n), 16)
			// now we can determine the length of the hexdecimal value and make an
			// offset for it
			blankSpace := hexValueOffset - len(buff)
			copy(s.fileBuff[blankSpace:], buff)
			copy(s.fileBuff[hexValueOffset:], crlf)
			copy(s.fileBuff[buffOffset+n:], crlf)

			if _, err := writer.Write(s.fileBuff[blankSpace : buffOffset+n+crlfSize]); err != nil {
				return status.ErrCloseConnection
			}
		}

		switch err {
		case nil:
		case io.EOF:
			_, err = writer.Write(chunkedFinalizer)
			return err
		default:
			return status.ErrCloseConnection
		}
	}
}

// renderHeaderInto the buffer. Appends CRLF in the end
func (s *serializer) renderHeader(header headers.Header) {
	s.buff = append(s.buff, header.Key...)
	s.colonsp()
	s.buff = append(s.buff, header.Value...)
	s.crlf()
}

func (s *serializer) renderCookie(c cookie.Cookie) {
	s.buff = append(s.buff, setCookie...)
	s.buff = append(s.buff, c.Name...)
	s.buff = append(s.buff, '=')
	s.buff = append(s.buff, c.Value...)
	s.buff = append(s.buff, ';', ' ')

	if len(c.Path) > 0 {
		s.buff = append(s.buff, "Path="...)
		s.buff = append(s.buff, c.Path...)
		s.buff = append(s.buff, ';', ' ')
	}

	if len(c.Domain) > 0 {
		s.buff = append(s.buff, "Domain="...)
		s.buff = append(s.buff, c.Domain...)
		s.buff = append(s.buff, ';', ' ')
	}

	if !c.Expires.IsZero() {
		s.buff = append(s.buff, "Expires="...)
		// TODO: this will probably be slow. Can be optimized via rendering it manually
		// TODO: directly into the s.buff
		s.buff = append(s.buff, c.Expires.In(gmt).Format(time.RFC1123)...)
		s.buff = append(s.buff, ';', ' ')
	}

	if c.MaxAge != 0 {
		maxage := "0"
		if c.MaxAge > 0 {
			maxage = strconv.Itoa(c.MaxAge)
		}

		s.buff = append(s.buff, "MaxAge="...)
		s.buff = append(s.buff, maxage...)
		s.buff = append(s.buff, ';', ' ')
	}

	if len(c.SameSite) > 0 {
		s.buff = append(s.buff, "SameSite="...)
		s.buff = append(s.buff, c.SameSite...)
		s.buff = append(s.buff, ';', ' ')
	}

	if c.Secure {
		s.buff = append(s.buff, "Secure; "...)
	}

	if c.HttpOnly {
		s.buff = append(s.buff, "HttpOnly; "...)
	}

	// strip last 2 bytes, which are always a semicolon and a space
	s.buff = s.buff[:len(s.buff)-2]
	s.crlf()
}

func (s *serializer) renderContentLength(value int64) {
	s.buff = strconv.AppendInt(append(s.buff, contentLength...), value, 10)
	s.crlf()
}

func (s *serializer) renderKnownHeader(key, value string) {
	s.buff = append(s.buff, key...)
	s.buff = append(s.buff, value...)
	s.crlf()
}

func (s *serializer) renderProtocol(protocol proto.Proto) {
	s.buff = append(s.buff, protocol.String()...)
}

func (s *serializer) sp() {
	s.buff = append(s.buff, ' ')
}

func (s *serializer) colonsp() {
	s.buff = append(s.buff, ':', ' ')
}

func (s *serializer) crlf() {
	s.buff = append(s.buff, crlf...)
}

func (s *serializer) clear() {
	s.buff = s.buff[:0]
	s.defaultHeaders.Reset()
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
	// used at initialization period only, so quite acceptable
	return fmt.Sprintf("%s: %s\r\n", key, value)
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
	for i := range d {
		d[i].Excluded = false
	}
}
