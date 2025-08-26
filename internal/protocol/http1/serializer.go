package http1

import (
	"io"
	"math/bits"
	"slices"
	"strconv"
	"strings"
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
	"github.com/indigo-web/indigo/internal/hexconv"
	"github.com/indigo-web/indigo/internal/response"
	"github.com/indigo-web/indigo/internal/strutil"
	"github.com/indigo-web/indigo/kv"
	"github.com/indigo-web/indigo/transport"
)

type serializer struct {
	cfg            *config.Config
	request        *http.Request
	response       *response.Fields
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
		cfg:     cfg,
		request: request,
		client:  client,
		codecs:  codecs,
		buff:    buff,
		defaultHeaders: newDefaultHeaders(
			pairsFromMap(cfg.Headers.Default, codecs.AcceptEncoding()),
		),
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
	resp := response.Expose()

	s.appendProtocol(protocol)
	s.appendStatus(resp)
	s.appendHeaders(resp)

	for _, c := range resp.Cookies {
		s.appendCookie(c)
	}

	err := s.writeStream(resp)
	if err != nil {
		return err
	}

	return s.flush()
}

func (s *serializer) writeStream(resp *response.Fields) (err error) {
	s.response = resp
	stream, length := resp.Stream, resp.StreamSize
	unsized := length == -1
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

	compression := resp.ContentEncoding
	if resp.AutoCompress && (unsized || length >= s.cfg.NET.SmallBody) {
		// if the stream is sized and the size is below limit (i.e. is considered a small one),
		// do not compress it. It won't give much gain anyway, yet the performance is impacted,
		// especially if we otherwise could use a zero-copy mechanism
		compression = s.request.PreferredEncoding()
	}

	compressor := s.getCompressor(compression)
	if !unsized && compressor != nil {
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

		// +len(crlf) because it wasn't written yet, therefore not yet included in the len(s.buff)
		s.growToContain(len(crlf) + int(length))
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

func (s *serializer) grow(newsize int) {
	// cap the size at its top value from the config.
	newsize = min(s.cfg.NET.WriteBufferSize.Maximal, newsize)
	// the growth can be triggered even the buffer is already at its maximal size. Do nothing then.
	if newsize > cap(s.buff) {
		s.buff = make([]byte, 0, newsize)
	}
}

func (s *serializer) growToContain(n int) {
	newsize := min(s.cfg.NET.WriteBufferSize.Maximal-len(s.buff), n)
	s.buff = slices.Grow(s.buff, newsize)
}

func (s *serializer) getCompressor(token string) codec.Compressor {
	if token == "" || token == "identity" {
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

	for i, header := range s.defaultHeaders {
		if header.Excluded {
			s.defaultHeaders[i].Excluded = false
			continue
		}

		s.appendHeader(header.Pair)
		s.crlf()
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

type chunkedWriter struct {
	s *serializer
}

func (c chunkedWriter) ReadFrom(r io.Reader) (total int64, err error) {
	const crlflen = len(crlf)

	for {
		var (
			buff       = c.s.buff[len(c.s.buff):cap(c.s.buff)]
			maxHexLen  = hexlen(len(buff))
			dataOffset = maxHexLen + crlflen
		)

		n, err := r.Read(buff[dataOffset : len(buff)-crlflen])
		if n > 0 {
			total += int64(n)

			if err := c.writechunk(maxHexLen, n); err != nil {
				return 0, err
			}

			if n+dataOffset+crlflen >= cap(c.s.buff)-cap(c.s.buff)>>6 {
				// if the chunk solely occupies ~98.44% of the whole buffer capacity, double the size
				c.s.grow(cap(c.s.buff) << 1)
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

	// knowing the size of b in advance, grow to contain it fully if needed
	c.s.grow(hexlen(cap(c.s.buff)) + crlflen + blen + crlflen + 1)

	for len(b) > 0 {
		var (
			buff       = c.s.buff[len(c.s.buff):cap(c.s.buff)]
			maxHexLen  = hexlen(len(buff))
			dataOffset = maxHexLen + crlflen
		)

		n = copy(buff[dataOffset:len(buff)-crlflen], b)
		if err = c.writechunk(maxHexLen, n); err != nil {
			return 0, err
		}

		b = b[n:]
	}

	return blen, nil
}

func (c chunkedWriter) writechunk(maxHexLen, datalen int) error {
	const crlflen = len(crlf)

	for {
		var (
			buff       = c.s.buff[len(c.s.buff):cap(c.s.buff)]
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

			continue
		}

		// write the zero-filled hex length
		chunklen := datalen
		for i := maxHexLen; i > 0; i-- {
			buff[i-1] = hexconv.Char[chunklen&0b1111]
			chunklen >>= 4
		}

		copy(buff[maxHexLen:], crlf)          // CRLF between length and data
		copy(buff[dataOffset+datalen:], crlf) // CRLF at the end of the data

		// extend the buffer to include the written data
		c.s.buff = c.s.buff[:len(c.s.buff)+dataOffset+datalen+crlflen]

		if !c.s.response.Buffered || len(c.s.buff) >= 3*cap(c.s.buff)>>2 {
			return c.s.flush()
		}

		return nil
	}
}

func (c chunkedWriter) Close() error {
	if err := c.s.safeAppend([]byte("0\r\n\r\n")); err != nil {
		return err
	}

	return c.s.flush()
}

type identityWriter struct {
	s *serializer
}

func (i identityWriter) ReadFrom(r io.Reader) (total int64, err error) {
	streamSize := i.s.response.StreamSize // guaranteed to be >0

	for total < streamSize {
		boundary := min(cap(i.s.buff), int(streamSize-total)+len(i.s.buff))
		n, err := r.Read(i.s.buff[len(i.s.buff):boundary])
		total += int64(n)

		i.s.buff = i.s.buff[0 : len(i.s.buff)+n]
		if !i.s.response.Buffered || len(i.s.buff) >= 3*cap(i.s.buff)>>2 {
			// flush if unbuffered OR buffered and the buffer is >=3/4 full.
			if ferr := i.s.flush(); ferr != nil {
				return 0, ferr
			}
		}

		switch err {
		case nil:
		case io.EOF:
			return total, nil
		default:
			return 0, err
		}
	}

	return total, nil
}

func (i identityWriter) Write(p []byte) (int, error) {
	return len(p), i.s.safeAppend(p)
}

func (i identityWriter) Close() error {
	return i.s.flush()
}

type excludablePair struct {
	Excluded bool
	kv.Pair
}

func pairsFromMap(m map[string]string, acceptEncoding string) []excludablePair {
	pairs := make([]excludablePair, 0, len(m)+1)
	pairs = append(pairs, excludablePair{
		Pair: kv.Pair{Key: "Accept-Encoding", Value: acceptEncoding},
	})

	for key, value := range m {
		pairs = append(pairs, excludablePair{
			Pair: kv.Pair{Key: key, Value: value},
		})
	}

	return pairs
}

type defaultHeaders []excludablePair

func newDefaultHeaders(pairs []excludablePair) defaultHeaders {
	slices.SortFunc(pairs, func(a, b excludablePair) int {
		return strings.Compare(a.Key, b.Key)
	})

	return pairs
}

func (d defaultHeaders) Exclude(key string) {
	// it's a perfect candidate for binary search, however in reality it introduced any visible
	// benefit only starting at 10 and more default headers. If less, the penalty is also significant.
	// The only optimization left to try out is stopping the iteration when `header.Key > key`,
	// considering the headers are still sorted.
	for i, header := range d {
		if strutil.CmpFoldFast(header.Key, key) {
			header.Excluded = true
			d[i] = header
			return
		}
	}
}

func hexlen(n int) int {
	return (bits.Len64(uint64(n))-1)>>2 + 1
}
