package http1

import (
	"github.com/indigo-web/indigo/config"
	"github.com/indigo-web/indigo/http"
	"github.com/indigo-web/indigo/http/cookie"
	"github.com/indigo-web/indigo/http/method"
	"github.com/indigo-web/indigo/http/proto"
	"github.com/indigo-web/indigo/http/status"
	"github.com/indigo-web/indigo/internal/response"
	"github.com/indigo-web/indigo/internal/strutil"
	"github.com/indigo-web/indigo/kv"
	"io"
	"math/bits"
	"net"
	"strconv"
	"strings"
	"time"
)

const crlf = "\r\n"

type writer interface {
	io.Writer
	Conn() net.Conn
}

type serializer struct {
	cfg            *config.Config
	request        *http.Request
	client         writer
	reader         constReader
	buff           []byte
	defaultHeaders defaultHeaders
}

func newSerializer(
	cfg *config.Config,
	client writer,
	buff []byte,
	defHdrs map[string]string,
	request *http.Request,
) *serializer {
	return &serializer{
		cfg:            cfg,
		request:        request,
		client:         client,
		buff:           buff,
		defaultHeaders: processDefaultHeaders(defHdrs),
	}
}

// Upgrade writes an informational response 101 Switching Protocols without immediately flushing it.
func (s *serializer) Upgrade() {
	s.appendProtocol(s.request.Protocol)
	s.buff = append(s.buff, "101 Switching Protocol\r\n"...)

	s.appendKnownHeader("Connection: ", "upgrade")
	s.appendKnownHeader("Upgrade: ", s.request.Upgrade.String())

	s.crlf()
}

func (s *serializer) Write(protocol proto.Protocol, response *http.Response) error {
	s.appendProtocol(protocol)
	resp := response.Reveal()
	s.appendResponseLine(resp)
	s.appendHeaders(resp)

	for _, c := range resp.Cookies {
		s.appendCookie(c)
	}

	err := s.writeStream(resp)
	if err != nil {
		return err
	}

	err = s.flush()

	if !isKeepAlive(protocol, s.request) && s.request.Upgrade == proto.Unknown {
		err = status.ErrCloseConnection
	}

	s.cleanup()
	return err
}

func (s *serializer) writeStream(resp *response.Fields) error {
	if bodyLen := len(resp.BufferedBody); bodyLen > 0 {
		s.reader.Reset(resp.BufferedBody)
		resp.Stream = &s.reader
		resp.StreamSize = int64(bodyLen)
	}

	// TODO: here we should grow the buffer if necessary

	switch {
	case resp.StreamSize > -1:
		s.appendContentLength(resp.StreamSize)
		s.crlf()
		if err := s.writePlainData(resp.Stream, resp.StreamSize); err != nil {
			return err
		}
	case resp.Stream != nil:
		s.appendKnownHeader("Transfer-Encoding: ", "chunked")
		s.crlf()
		if err := s.writeChunked(resp.Stream); err != nil {
			return err
		}
	default:
		s.appendKnownHeader("Content-Length: ", "0")
		s.crlf()
	}

	if c, ok := resp.Stream.(io.Closer); ok {
		return c.Close()
	}

	return nil
}

func (s *serializer) writePlainData(r io.Reader, size int64) error {
	if size == 0 || s.request.Method == method.HEAD {
		return nil
	}

	if r == nil {
		// FIXME: this is clearly the user's fault. And this must somehow be signalized
		// FIXME: explicitly, otherwise debugging this shit might become a personal hell.
		return status.ErrInternalServerError
	}

	if w, ok := r.(io.WriterTo); ok {
		// files do support the io.WriterTo interface and take advantage of smarter kernel
		// mechanisms if available. In Linux it's sendfile(2), for example. File entities are
		// smart and know if the passed Writer is a net.Conn
		if err := s.flush(); err != nil {
			return err
		}

		_, err := w.WriteTo(s.client.Conn())
		return err
	}

	for size > 0 {
		boundary := min(int64(len(s.buff))+size, int64(cap(s.buff)))
		n, err := r.Read(s.buff[len(s.buff):boundary])
		size -= int64(n)

		if size > 0 && err != nil {
			if err == io.EOF {
				err = io.ErrUnexpectedEOF
			}

			return err
		}

		s.buff = s.buff[:len(s.buff)+n]
		if err = s.flush(); err != nil {
			return err
		}
	}

	return s.flush()
}

var (
	// chunkExtZeroFill is used to fill the gap between chunk length and chunk content. The count
	// 64/4 represents 64 bits - the maximal uint size, and 4 - bits per hex value, therefore
	// resulting in 15 characters (plus semicolon) total.
	chunkExtZeroFill = ";" + strings.Repeat("0", 64/4-1)
	chunkZeroTrailer = []byte("0\r\n\r\n")
)

func (s *serializer) writeChunked(r io.Reader) error {
	if r == nil {
		// TODO: the reader should not be nil at all. This is a probable error, therefore
		// TODO: must somehow be signalized to the user.
		return s.safeAppend(chunkZeroTrailer)
	}

	if s.request.Method == method.HEAD {
		return nil
	}

	const crlflen = len(crlf)

	for {
		var (
			buff         = s.buff[len(s.buff):cap(s.buff)]
			maxHexLength = (bits.Len64(uint64(len(buff)))-1)/4 + 1
			dataOffset   = maxHexLength + crlflen
		)

		if len(buff) <= dataOffset+crlflen {
			// FIXME: in case the response buffer is too small, this will cause an infinite loop.
			// FIXME: Consider checking it before and panicking (?)
			if err := s.flush(); err != nil {
				return err
			}

			continue
		}

		n, err := r.Read(buff[dataOffset : len(buff)-crlflen])

		hexlen := len(strconv.AppendUint(buff[:0], uint64(n), 16)) // chunk length
		copy(buff[hexlen:maxHexLength], chunkExtZeroFill)          // fill gap between length and (future) CRLF
		copy(buff[maxHexLength:], crlf)                            // CRLF between length and data
		copy(buff[dataOffset+n:], crlf)                            // CRLF at the end of the data

		s.buff = s.buff[:len(s.buff)+dataOffset+n+crlflen] // extend buffer to include the written data

		switch err {
		case nil:
		case io.EOF:
			if n != 0 {
				if err = s.safeAppend(chunkZeroTrailer); err != nil {
					return err
				}
			}

			return s.flush()
		default:
			return err
		}

		if err = s.flush(); err != nil {
			return err
		}
	}
}

// safeAppend tries to append a string into a limited capacity buffer, which can possibly overflow.
// If the input data is longer than free space left in the buffer, the buffer is filled till full
// and flushed, leaving thereby free space for the rest of the string.
func (s *serializer) safeAppend(data []byte) error {
	for len(data) > 0 {
		freeSpace := cap(s.buff) - len(s.buff)

		if len(data) <= freeSpace {
			s.buff = append(s.buff, data...)
			return nil
		}

		s.buff = append(s.buff, data[:freeSpace]...)
		if err := s.flush(); err != nil {
			return err
		}

		data = data[freeSpace:]
	}

	return nil
}

func (s *serializer) flush() (err error) {
	if len(s.buff) > 0 {
		_, err = s.client.Write(s.buff)
		s.buff = s.buff[:0]
	}

	return err
}

func (s *serializer) appendResponseLine(fields *response.Fields) {
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

func (s *serializer) appendHeaders(fields *response.Fields) {
	responseHeaders := fields.Headers

	for _, header := range responseHeaders {
		s.appendHeader(header)
		s.defaultHeaders.Exclude(header.Key)
	}

	for _, header := range s.defaultHeaders {
		if header.Excluded {
			continue
		}

		s.buff = append(s.buff, header.Full...)
	}

	s.appendKnownHeader("Content-Type: ", fields.ContentType)
}

// appendHeader writes a complete header field line, including the crlf at the end.
func (s *serializer) appendHeader(header kv.Pair) {
	s.buff = append(s.buff, header.Key...)
	s.colonsp()
	s.buff = append(s.buff, header.Value...)
	s.crlf()
}

// appendKnownHeader differs from appendHeader only by the fact that the key is known to already
// have a colon and a space included.
func (s *serializer) appendKnownHeader(key, value string) {
	s.buff = append(s.buff, key...)
	s.buff = append(s.buff, value...)
	s.crlf()
}

var zoneGMT = time.FixedZone("GMT", 0)

func (s *serializer) appendCookie(c cookie.Cookie) {
	s.buff = append(s.buff, "Set-Cookie: "...)
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
		// TODO: this might be slow. We can write the date manually though
		s.buff = c.Expires.In(zoneGMT).AppendFormat(s.buff, time.RFC1123)
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

func (s *serializer) appendContentLength(value int64) {
	s.buff = append(s.buff, "Content-Length: "...)
	s.buff = strconv.AppendUint(s.buff, uint64(value), 10)
	s.crlf()
}

func (s *serializer) appendProtocol(protocol proto.Protocol) {
	// in case the request method or path were malformed, parser had no chance of reaching
	// the protocol and thereby resulting in the unknown one.
	const defaultFallbackProtocol = proto.HTTP11

	if protocol == proto.Unknown {
		protocol = defaultFallbackProtocol
	}

	s.buff = append(s.buff, protocol.String()...)
	s.sp()
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

func (s *serializer) cleanup() {
	s.defaultHeaders.Reset()
}

func isKeepAlive(protocol proto.Protocol, req *http.Request) bool {
	switch protocol {
	case proto.HTTP10:
		return strutil.CmpFold(req.Connection, "keep-alive")
	case proto.HTTP11:
		// in case of HTTP/1.1, keep-alive may be only disabled
		return !strutil.CmpFold(req.Connection, "close")
	default:
		// as the protocol is unknown and the code was probably caused by some sort
		// of bug, consider closing it
		return false
	}
}

func processDefaultHeaders(hdrs map[string]string) defaultHeaders {
	processed := make(defaultHeaders, 0, len(hdrs))

	for key, value := range hdrs {
		full := key + ": " + value + crlf
		processed = append(processed, defaultHeader{
			// we let the GC release all the values of the map, as here we're using only
			// the brand-new line without keeping the original string
			Key:  full[:len(key)],
			Full: full,
		})
	}

	return processed
}

type constReader struct {
	data []byte
}

func (c *constReader) Read(b []byte) (n int, err error) {
	n = copy(b, c.data)
	c.data = c.data[n:]
	if len(c.data) == 0 {
		err = io.EOF
	}

	return n, err
}

func (c *constReader) Reset(data []byte) {
	c.data = data
}

type defaultHeader struct {
	Excluded bool
	Key      string
	Full     string
}

type defaultHeaders []defaultHeader

func (d defaultHeaders) Exclude(key string) {
	for i, header := range d {
		if strutil.CmpFold(header.Key, key) {
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
