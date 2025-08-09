package http1

import (
	"io"
	"math/bits"
	"slices"
	"strconv"
	"time"

	"github.com/indigo-web/indigo/config"
	"github.com/indigo-web/indigo/http"
	"github.com/indigo-web/indigo/http/codec"
	"github.com/indigo-web/indigo/http/cookie"
	"github.com/indigo-web/indigo/http/method"
	"github.com/indigo-web/indigo/http/mime"
	"github.com/indigo-web/indigo/http/proto"
	"github.com/indigo-web/indigo/http/status"
	"github.com/indigo-web/indigo/internal/codecutil"
	"github.com/indigo-web/indigo/internal/response"
	"github.com/indigo-web/indigo/internal/strutil"
	"github.com/indigo-web/indigo/kv"
	"github.com/indigo-web/indigo/transport"
)

type serializer struct {
	cfg            *config.Config
	request        *http.Request
	client         transport.Client
	buff           []byte
	streamReadBuff []byte
	defaultHeaders defaultHeaders
	codecs         codecutil.Cache
}

func newSerializer(
	cfg *config.Config,
	request *http.Request,
	client transport.Client,
	codecs codecutil.Cache,
	buff []byte,
) *serializer {
	return &serializer{
		cfg:            cfg,
		request:        request,
		client:         client,
		codecs:         codecs,
		buff:           buff,
		defaultHeaders: preprocessDefaultHeaders(cfg.Headers.Default, codecs.AcceptEncodings()),
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
	resp := response.Expose()
	s.appendStatus(resp)
	s.appendHeaders(resp)

	for _, c := range resp.Cookies {
		s.appendCookie(c)
	}

	err := s.writeStream(resp)
	if err != nil {
		return err
	}

	err = s.flush()
	s.cleanup()

	return err
}

func (s *serializer) writeStream(resp *response.Fields) (err error) {
	stream, length := resp.Stream, resp.StreamSize
	if length == 0 {
		s.appendKnownHeader("Content-Length: ", "0")
		s.crlf()
		return nil
	}

	if stream == nil {
		// TODO: add debug mode, in which errors caused by the user are described in details in the response body
		return status.ErrInternalServerError
	}

	defer func() {
		if c, ok := stream.(io.Closer); ok {
			if cerr := c.Close(); cerr != nil && err == nil {
				err = cerr
			}
		}
	}()

	var encoder io.WriteCloser

	compressor := s.getCompressor(resp.ContentEncoding)
	if length != -1 && compressor != nil {
		// if sized stream is compressed, convert it to unsized
		length = -1
	}

	if length == -1 {
		encoder = chunkedWriter{s}
		s.appendKnownHeader("Transfer-Encoding: ", "chunked")
	} else {
		encoder = identityWriter{s}
		s.appendContentLength(length)

		if wt, ok := stream.(io.WriterTo); ok && s.request.Method != method.HEAD {
			// there are chances to engage some smarter ways to transfer the stream.
			// For example, sendfile(2) on files when running on Linux.

			s.crlf() // to finalize the headers block

			if err = s.flush(); err != nil {
				return err
			}

			_, err = wt.WriteTo(s.client.Conn())
			return err
		}

		s.growToContain(int(length) + len(crlf)) // because CRLF wasn't written yet
	}

	s.crlf() // finalize the headers block

	if s.request.Method == method.HEAD {
		return nil
	}

	if compressor != nil {
		compressor.ResetCompressor(encoder)
		encoder = compressor
	}

	defer func() {
		if cerr := encoder.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()

	if rf, ok := encoder.(io.ReaderFrom); ok {
		_, err = rf.ReadFrom(stream)
		return err
	}

	if wt, ok := stream.(io.WriterTo); ok {
		_, err = wt.WriteTo(encoder)
		return err
	}

	for {
		// TODO: if we use a big slice split in half for each buff and streamReadBuff, we could
		// TODO: get by with just a single slices.Grow() call.
		if cap(s.buff) > cap(s.streamReadBuff) {
			s.streamReadBuff = slices.Grow(s.streamReadBuff[:0], cap(s.buff))
		}

		n, err := stream.Read(s.streamReadBuff[:cap(s.streamReadBuff)])
		if n > 0 {
			_, encerr := encoder.Write(s.streamReadBuff[:n])
			if encerr != nil {
				return encerr
			}
		}

		switch err {
		case nil:
		case io.EOF:
			return nil
		default:
			return err
		}
	}
}

func (s *serializer) growToContain(n int) {
	extra := min(s.cfg.NET.WriteBufferSize.Maximal-cap(s.buff), n)
	if extra <= 0 {
		// surplus is possible and in fact isn't a bug, because slices.Grow doesn't guarantee
		// that the grown slice will have a specific capacity. It is guaranteed to be _enough_
		// to hold n new values. It's anyway not that bad to have a few extra bytes, after all.
		return
	}

	s.buff = slices.Grow(s.buff, extra)
}

func (s *serializer) getCompressor(token string) codec.Compressor {
	if len(token) == 0 {
		return nil
	}

	compressor := s.codecs.Get(token)
	if compressor != nil {
		s.appendKnownHeader("Content-Encoding: ", token)
	}

	return compressor
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

func (s *serializer) appendStatus(fields *response.Fields) {
	if code := status.StringCode(fields.Code); len(code) > 0 {
		s.buff = append(s.buff, code...)
	} else {
		// some non-standard code
		s.buff = strconv.AppendUint(s.buff, uint64(fields.Code), 10)
	}

	s.sp()

	statusText := fields.Status
	if len(statusText) == 0 {
		statusText = status.FromCode(fields.Code)
	}

	s.buff = append(s.buff, statusText...)
	s.crlf()
}

func (s *serializer) appendHeaders(fields *response.Fields) {
	responseHeaders := fields.Headers

	for _, header := range responseHeaders {
		s.defaultHeaders.Exclude(header.Key)
		s.appendHeader(header)

		if strutil.CmpFoldFast(header.Key, "Content-Type") && fields.Charset != mime.Unset {
			s.buff = append(s.buff, "; charset="...)
			s.buff = append(s.buff, fields.Charset...)
		}

		s.crlf()
	}

	for _, header := range s.defaultHeaders {
		if header.Excluded {
			continue
		}

		s.buff = append(s.buff, header.Full...)
	}
}

// appendHeader writes a complete header field line excluding the trailing CRLF.
func (s *serializer) appendHeader(header kv.Pair) {
	s.buff = append(s.buff, header.Key...)
	s.colonsp()
	s.buff = append(s.buff, header.Value...)
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
		// TODO: this _may_ be slow. We could write it manually instead
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
	if protocol == proto.Unknown {
		// in case the request method or path were malformed, parser had no chance of reaching
		// the protocol and thereby resulting in the unknown one.
		protocol = proto.HTTP11
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

const crlf = "\r\n"

func (s *serializer) crlf() {
	s.buff = append(s.buff, crlf...)
}

func (s *serializer) cleanup() {
	s.defaultHeaders.Reset()
}

type chunkedWriter struct {
	s *serializer
}

func (c chunkedWriter) maxhex(n int) int {
	return (bits.Len64(uint64(n))-1)>>2 + 1
}

func (c chunkedWriter) ReadFrom(r io.Reader) (total int64, err error) {
	const crlflen = len(crlf)

	for {
		var (
			buff         = c.s.buff[len(c.s.buff):cap(c.s.buff)]
			maxHexLength = c.maxhex(len(buff))
			dataOffset   = maxHexLength + crlflen
		)

		n, err := r.Read(buff[dataOffset : len(buff)-crlflen])
		if n > 0 {
			total += int64(n)

			if err := c.writechunk(maxHexLength, n); err != nil {
				return 0, err
			}
		}

		switch err {
		case nil:
			continue
		case io.EOF:
			return total, nil
		default:
			return 0, err
		}
	}
}

func (c chunkedWriter) Write(b []byte) (n int, err error) {
	const crlflen = len(crlf)
	blen := len(b)

	// TODO: add optional but enabled by default buffering chunks.
	// otherwise, cap(b) = cap(c.s.buff) => 7 bytes leftover, which will be sent
	// as an independent chunk. Highly inefficient. However, making buffering a default behaviour
	// completely disables the possibility to implement longpolling based on chunked transfer encoding.
	// Also undesired. But by default very little people do really use it like that or anyhow rely on
	// lag-free chunk upload or on their consistency. Therefore, enabling buffering by default would result
	// in a great choice.

	for len(b) > 0 {
		buff := c.s.buff[len(c.s.buff):cap(c.s.buff)]
		maxHexLen := c.maxhex(len(buff))

		n = copy(buff[maxHexLen+crlflen:len(buff)-crlflen], b)
		if err = c.writechunk(maxHexLen, n); err != nil {
			return 0, err
		}

		b = b[n:]
	}

	return blen, nil
}

func (c chunkedWriter) writechunk(maxHexLen, datalen int) error {
	const crlflen = len(crlf)

	// TODO: add optional but enabled by default buffering chunks.
	// otherwise, cap(b) = cap(c.s.buff) => 7 bytes leftover, which will be sent
	// as an independent chunk. Highly inefficient. However, making buffering a default behaviour
	// completely disables the possibility to implement longpolling based on chunked transfer encoding.
	// Also undesired. But by default very little people do really use it like that or anyhow rely on
	// lag-free chunk upload or on their consistency. Therefore, enabling buffering by default would result
	// in a great choice.

	for {
		var (
			buff       = c.s.buff[len(c.s.buff):cap(c.s.buff)]
			buffOffset = 0
			dataOffset = maxHexLen + crlflen
		)

		if len(buff) <= dataOffset+crlflen {
			// this is normally caused when headers took up almost all available buffer space.
			if cap(buff) <= dataOffset+crlflen {
				// but also might if the buffer is itself way too small, even if we completely
				// clean it. In practice this can only happen in tests, because otherwise the
				// buffer is naturally grown to at least 16-64 bytes because of a response line
				// and inevitable headers, like Content-Length and Accept-Encoding
				return status.ErrInternalServerError
			}

			if err := c.s.flush(); err != nil {
				return err
			}

			// TODO: buffer chunks here
			continue
		}

		hexlen := len(strconv.AppendUint(buff[:0], uint64(datalen), 16)) // chunk length

		if len(c.s.buff) > 0 {
			// if there was any data in the buffer before, we must fill the gap in between.
			// The best way to do it is via an extension.
			copy(buff[hexlen:maxHexLen], chunkExtZeroFill)
		} else {
			// otherwise, we can save a couple of bytes by simply truncating the unused prefix slots.
			buffOffset = maxHexLen - hexlen
			copy(buff[buffOffset:], buff[:hexlen])
		}

		copy(buff[maxHexLen:], crlf)          // CRLF between length and data
		copy(buff[dataOffset+datalen:], crlf) // CRLF at the end of the data

		restore := c.s.buff[:0]
		c.s.buff = c.s.buff[buffOffset : len(c.s.buff)+dataOffset+datalen+crlflen] // extend buffer to include the written data
		if err := c.s.flush(); err != nil {
			return err
		}
		c.s.buff = restore

		if cap(c.s.buff)-dataOffset-datalen-crlflen <= cap(c.s.buff)>>6 {
			// if free space left after the whole chunk was written is less than
			// ~1.56% of the buffer total capacity, double the buffer size.
			newsize := min(c.s.cfg.NET.WriteBufferSize.Maximal, cap(c.s.buff)<<1)
			// the growth can be triggered even the buffer is already at its maximal size. Do nothing then.
			if newsize > cap(c.s.buff) {
				c.s.buff = make([]byte, 0, newsize)
			}
		}

		return nil
	}
}

func (c chunkedWriter) Close() error {
	if err := c.s.safeAppend(chunkZeroTrailer); err != nil {
		return err
	}

	return c.s.flush()
}

type identityWriter struct {
	s *serializer
}

func (i identityWriter) ReadFrom(r io.Reader) (total int64, err error) {
	for {
		n, err := r.Read(i.s.buff[len(i.s.buff):cap(i.s.buff)])
		total += int64(n)
		i.s.buff = i.s.buff[0 : len(i.s.buff)+n]

		if err := i.s.flush(); err != nil {
			return 0, err
		}

		switch err {
		case nil:
		case io.EOF:
			return total, nil
		default:
			return 0, err
		}
	}
}

func (i identityWriter) Write(p []byte) (int, error) {
	err := i.s.safeAppend(p)
	return len(p), err
}

func (i identityWriter) Close() error {
	return i.s.flush()
}

func preprocessDefaultHeaders(headers map[string]string, acceptEncoding string) defaultHeaders {
	processed := make(defaultHeaders, 0, len(headers)+1)

	for key, value := range headers {
		serialized := key + ": " + value + crlf
		processed = append(processed, defaultHeader{
			// we let the GC release all the values of the map, as here we're using only
			// the brand-new line without keeping the original string
			Key:  serialized[:len(key)],
			Full: serialized,
		})
	}

	processed = append(processed, defaultHeader{
		Key:  "Accept-Encoding",
		Full: "Accept-Encoding: " + acceptEncoding + crlf,
	})

	return processed
}

type defaultHeader struct {
	Excluded bool
	Key      string
	Full     string
}

type defaultHeaders []defaultHeader

func (d defaultHeaders) Exclude(key string) {
	// TODO: binary search

	for i, header := range d {
		if strutil.CmpFoldFast(header.Key, key) {
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
