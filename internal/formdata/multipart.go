package formdata

import (
	"github.com/indigo-web/indigo/config"
	"github.com/indigo-web/indigo/http/form"
	"github.com/indigo-web/indigo/http/status"
	"github.com/indigo-web/indigo/internal/strutil"
	"github.com/indigo-web/indigo/internal/urlencoded"
	"github.com/indigo-web/utils/uf"
	"iter"
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
	s := newStream(uf.B2S(data))

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

		// TODO: DecodeString doesn't decode plus-symbols as spaces, but we need it

		var err error
		hdr.Name, buff, err = urlencoded.ExtendedDecodeString(hdr.Name, buff)
		if err != nil {
			return nil, err
		}

		hdr.File, buff, err = urlencoded.ExtendedDecodeString(hdr.File, buff)
		if err != nil {
			return nil, err
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
