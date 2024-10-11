package internal

import (
	"github.com/indigo-web/indigo/http/form"
	"github.com/indigo-web/indigo/http/headers"
	"github.com/indigo-web/indigo/http/mime"
	"github.com/indigo-web/indigo/http/status"
	"github.com/indigo-web/utils/uf"
	"iter"
)

const (
	DefaultCoding      = "utf8"
	DefaultContentType = mime.Plain
)

type header struct {
	Ok                               bool
	Name, File, ContentType, Charset string
}

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
	s := newStream(uf.B2S(data))

	if !skipPreamble(&s, boundary) {
		return nil, status.ErrBadRequest
	}

	if !s.Consume("\r\n") {
		return nil, status.ErrBadRequest
	}

	for hdr, value := range formParts(&s, boundary) {
		if !hdr.Ok || len(hdr.Name) == 0 {
			return nil, status.ErrBadRequest
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
			hdr.ContentType = DefaultContentType
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

func parseHeaders(s *stream) (header header) {
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
		// TODO: I assume there can be parameters I should look after
		origin.ContentType, ok = s.AdvanceLine()
		if !ok {
			return origin, false
		}

		var params string
		origin.ContentType, params = headers.CutParams(origin.ContentType)
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
	for key, value := range headers.WalkParams(params) {
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
	for key, value := range headers.WalkParams(params) {
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
