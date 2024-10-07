package formdata

import (
	"github.com/indigo-web/indigo/http/form"
	"github.com/indigo-web/indigo/http/status"
	"github.com/indigo-web/indigo/internal/formdata/internal"
	"github.com/indigo-web/utils/uf"
	"iter"
	"strings"
)

const DefaultCoding = "utf8"

func ParseMultipart(into form.Form, data []byte, b string) (form.Form, error) {
	var boundary string
	const boundaryFraming = len("--") + len("\r\n")

	if len(b) < 512-boundaryFraming {
		tmp := [512]byte{0: '-', 1: '-'}
		boundary = uf.B2S(tmp[:copy(tmp[2:], b)+2])
	} else {
		boundary = "--" + b
	}

	charset := DefaultCoding
	s := internal.NewStream(uf.B2S(data))

	if !stripPreamble(&s, boundary) {
		return nil, status.ErrBadRequest
	}

	if !s.Consume("\r\n") {
		return nil, status.ErrBadRequest
	}

	for hdr, value := range formParts(&s, boundary) {
		if !hdr.Ok {
			return nil, status.ErrBadRequest
		}

		if hdr.Name == "_charset_" {
			charset = hdr.Charset
		}

		if len(hdr.Charset) == 0 {
			hdr.Charset = charset
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

type header struct {
	Ok                               bool
	Name, File, ContentType, Charset string
}

func stripPreamble(s *internal.Stream, boundary string) bool {
	b := s.FindSubstr(boundary)
	if b == -1 {
		return false
	}

	s.Advance(b + len(boundary))
	return true
}

func formParts(s *internal.Stream, boundary string) iter.Seq2[header, string] {
	return func(yield func(header, string) bool) {
		for {
			hdr := parseHeaders(s)
			if !hdr.Ok {
				yield(hdr, "")
				return
			}

			next := s.FindSubstr(boundary)
			if next == -1 {
				yield(header{Ok: false}, "")
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

func parseHeaders(s *internal.Stream) (header header) {
	for {
		var ok bool
		header, ok = parseHeader(s, header)
		header.Ok = ok

		if !ok {
			return header
		}

		if s.Consume("\r\n") {
			return header
		}
	}
}

func parseHeader(s *internal.Stream, origin header) (modified header, ok bool) {
	switch {
	case s.ConsumeFold("Content-Disposition:"):
		s.SkipWhitespaces()
		s.Consume("form-data;")
		s.SkipWhitespaces()
		params, ok := s.AdvanceLine()
		if !ok {
			return origin, false
		}

		return parseCDParams(params, origin)
	case s.ConsumeFold("Content-Type:"):
		s.SkipWhitespaces()
		// TODO: I assume there can be parameters I should look after
		origin.ContentType, ok = s.AdvanceLine()

		return origin, ok
	default:
		// must ignore
		_, ok = s.AdvanceLine()
		return origin, ok
	}
}

func parseCDParams(params string, origin header) (modified header, ok bool) {
	for {
		key, other, found := cutByte(params, '=')
		if !found {
			return origin, len(stripWS(key)) == 0
		}

		var value string
		value, params, found = cutByte(other, ';')

		switch key {
		case "name":
			origin.Name = value
		case "filename":
			origin.File = value
			// others must ignore
		}

		if !found {
			return origin, true
		}
	}
}

func cutByte(str string, c byte) (before string, after string, found bool) {
	pos := strings.IndexByte(str, c)
	if pos == -1 {
		return str, "", false
	}

	return str[:pos], str[pos+1:], true
}

func stripWS(str string) string {
	for i, c := range str {
		switch c {
		case ' ', '\t':
		default:
			return str[i:]
		}
	}

	return ""
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
