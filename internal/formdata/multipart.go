package formdata

import (
	"github.com/flrdv/uf"
	"github.com/indigo-web/indigo/config"
	"github.com/indigo-web/indigo/http/form"
	"github.com/indigo-web/indigo/http/status"
	"github.com/indigo-web/indigo/internal/hexconv"
	"github.com/indigo-web/indigo/internal/strutil"
	"iter"
	"strings"
)

type header struct {
	Name, File, ContentType, Charset string
}

func ParseMultipart(cfg *config.Config, into form.Form, data, buff []byte, b string) (form.Form, error) {
	var boundary string
	const boundaryFraming = len("--") + len("\r\n")

	if len(b) < 512-boundaryFraming {
		tmp := [512]byte{0: '-', 1: '-'}
		boundary = uf.B2S(tmp[:copy(tmp[2:], b)+2])
	} else {
		boundary = "--" + b
	}

	charset := cfg.Body.Form.DefaultCoding
	s := stream(uf.B2S(data))

	if !skipPreamble(&s, boundary) {
		return nil, status.ErrBadRequest
	}

	if !s.Consume("\r\n") {
		return nil, status.ErrBadRequest
	}

	for hdr, value := range formParts(&s, boundary) {
		if len(hdr.Name) == 0 {
			return nil, status.ErrBadRequest
		}

		var ok bool
		hdr.Name, buff, ok = urldecode(hdr.Name, buff)
		if !ok {
			return nil, status.ErrBadEncoding
		}

		hdr.File, buff, ok = urldecode(hdr.File, buff)
		if !ok {
			return nil, status.ErrBadEncoding
		}

		if hdr.Name == "_charset_" {
			charset = value
			if len(charset) == 0 {
				return nil, status.ErrBadRequest
			}

			continue
		}

		if len(hdr.Charset) == 0 {
			hdr.Charset = charset
		}

		if len(hdr.ContentType) == 0 {
			hdr.ContentType = cfg.Body.Form.DefaultContentType
		}

		into = append(into, form.Data{
			Name:     hdr.Name,
			Filename: hdr.File,
			Type:     hdr.ContentType,
			Charset:  hdr.Charset,
			Value:    value,
		})
	}

	return into, nil
}

func skipPreamble(s *stream, boundary string) bool {
	b := s.FindSubstr(boundary)
	if b == -1 {
		return false
	}

	s.Advance(b + len(boundary))
	return true
}

func formParts(s *stream, boundary string) iter.Seq2[header, string] {
	return func(yield func(header, string) bool) {
		for {
			hdr := parseHeaders(s)
			if len(hdr.Name) == 0 {
				yield(hdr, "")
				return
			}

			next := s.FindSubstr(boundary)
			if next == -1 {
				yield(header{}, "")
				return
			}

			if !yield(hdr, rstripCRLF(s.Advance(next))) {
				return
			}

			s.Advance(len(boundary))

			if s.Consume("--\r\n") {
				return
			}
		}
	}
}

func urldecode(value string, buff []byte) (string, []byte, bool) {
	escaped := false
	checkpoint := 0
	offset := len(buff)

	for i := 0; i < len(value); i++ {
		switch value[i] {
		case '+':
			// it would be better to mark the string as escaped here and do all the same just as
			// in the case below, however, well... Let it be our tiny trick.
			bytes := uf.S2B(value)
			bytes[i] = ' '
		case '%':
			escaped = true
			if i+2 >= len(value) {
				return "", nil, false
			}

			buff = append(buff, value[checkpoint:i]...)
			a, b := hexconv.Halfbyte[value[i+1]], hexconv.Halfbyte[value[i+2]]
			if a|b == 0xFF {
				return "", nil, false
			}

			buff = append(buff, (a<<4)|b)
			i += 2
			checkpoint = i + 1
		}
	}

	if escaped {
		buff = append(buff, value[checkpoint:]...)

		return uf.B2S(buff[offset:]), buff, true
	}

	return value, buff, true
}

func parseHeaders(s *stream) (hdr header) {
	for {
		var ok bool
		hdr, ok = parseHeader(s, hdr)
		if !ok {
			return header{}
		}

		if s.Consume("\r\n") {
			return hdr
		}
	}
}

func parseHeader(s *stream, origin header) (modified header, ok bool) {
	switch {
	case s.ConsumeFold("Content-Disposition:"):
		s.SkipWhitespaces()
		s.Consume("form-data;")
		s.SkipWhitespaces()
		params, ok := s.AdvanceLine()
		if !ok {
			return origin, false
		}

		return parseContentDispositionParams(params, origin)
	case s.ConsumeFold("Content-Type:"):
		s.SkipWhitespaces()
		// TODO: there are probably important parameters I should take care of, but...
		origin.ContentType, ok = s.AdvanceLine()
		if !ok {
			return origin, false
		}

		var params string
		origin.ContentType, params = strutil.CutHeader(origin.ContentType)
		if len(params) > 0 {
			origin, ok = parseContentTypeParams(params, origin)
		}

		return origin, ok
	default:
		// must ignore
		_, ok = s.AdvanceLine()
		return origin, ok
	}
}

func parseContentDispositionParams(params string, origin header) (modified header, ok bool) {
	for key, value := range strutil.WalkKV(params) {
		if len(key) == 0 || len(value) == 0 {
			return origin, false
		}

		switch key {
		case "name":
			origin.Name = value
		case "filename":
			origin.File = value
		}
	}

	return origin, true
}

func parseContentTypeParams(params string, origin header) (modified header, ok bool) {
	for key, value := range strutil.WalkKV(params) {
		if len(key) == 0 || len(value) == 0 {
			return origin, false
		}

		if key == "charset" {
			origin.Charset = value
			return origin, true
		}
	}

	return origin, true
}

func rstripCRLF(str string) string {
	if str[len(str)-1] == '\n' {
		str = str[:len(str)-1]

		if str[len(str)-1] == '\r' {
			str = str[:len(str)-1]
		}
	}

	return str
}

type stream string

func (s *stream) Find(char byte) int {
	return strings.IndexByte(string(*s), char)
}

func (s *stream) FindSubstr(str string) int {
	for {
		begin := s.Find(str[0])
		if begin == -1 {
			return -1
		}

		if s.Compare(begin, str) {
			return begin
		}

		s.Advance(1)
	}
}

func (s *stream) Compare(offset int, str string) bool {
	if len(*s) < len(str)+offset {
		return false
	}

	return string(*s)[offset:offset+len(str)] == str
}

func (s *stream) CompareFold(offset int, str string) bool {
	if len(*s) < len(str)+offset {
		return false
	}

	return strutil.CmpFold(string(*s)[offset:offset+len(str)], str)
}

func (s *stream) Consume(str string) bool {
	if s.Compare(0, str) {
		s.Advance(len(str))
		return true
	}

	return false
}

func (s *stream) ConsumeFold(str string) bool {
	if s.CompareFold(0, str) {
		s.Advance(len(str))
		return true
	}

	return false
}

func (s *stream) Advance(n int) (leftBehind string) {
	leftBehind = string(*s)[:n]
	*s = stream(string(*s)[n:])
	return leftBehind
}

func (s *stream) AdvanceExclusively(n int) (leftBehind string) {
	leftBehind = s.Advance(n + 1)
	return leftBehind[:len(leftBehind)-1]
}

func (s *stream) AdvanceLine() (leftBehind string, ok bool) {
	newline := s.Find('\n')
	if newline == -1 {
		return "", false
	}

	leftBehind = s.AdvanceExclusively(newline)
	if leftBehind[len(leftBehind)-1] == '\r' {
		return leftBehind[:len(leftBehind)-1], true
	}

	return leftBehind, true
}

func (s *stream) SkipWhitespaces() {
	*s = stream(strutil.LStripWS(string(*s)))
}
