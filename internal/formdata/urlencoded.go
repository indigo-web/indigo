package formdata

import (
	"github.com/flrdv/uf"
	"github.com/indigo-web/indigo/http/form"
	"github.com/indigo-web/indigo/http/status"
	"github.com/indigo-web/indigo/internal/hexconv"
)

// ParseFormURLEncoded implements parser for key-value pairs similar to request path parameters. The only
// difference is that this parser is greedy, as is intended to parse an already received body.
//
// It prohibits any non-printable characters in keys but allows such in values, therefore printing them
// might cause unwanted effects (e.g. beeping terminal.)
func ParseFormURLEncoded(into form.Form, data, buff []byte) (result form.Form, buffer []byte, err error) {
	var (
		key     string
		escaped bool
		offset  int
		buffptr int
	)

parseKey:
	if len(data) == 0 {
		return into, buff, nil
	}

	buffptr = len(buff)

	for i := 0; i < len(data); i++ {
		switch char := data[i]; char {
		case '=':
			if escaped {
				buff = append(buff, data[offset:i]...)
				key = uf.B2S(buff[buffptr:])
				escaped = false
			} else {
				key = uf.B2S(data[:i])
			}

			data = data[i+1:]
			offset = 0
			goto parseValue
		case '&':
			if escaped {
				buff = append(buff, data[offset:i]...)
				key = uf.B2S(buff[buffptr:])
				escaped = false
			} else {
				key = uf.B2S(data[:i])
			}

			if len(key) == 0 {
				return into, buff, status.ErrBadEncoding
			}

			into = append(into, form.Data{Name: key})

			data = data[i+1:]
			offset = 0
			goto parseKey
		case '+':
			// not the thing I'm proud of, to be honest. I'd rather make here the same
			// as with percent-encoded characters, but...
			data[i] = ' '
		case '%':
			escaped = true
			if i+2 >= len(data) {
				return into, buff, status.ErrBadEncoding
			}

			buff = append(buff, data[offset:i]...)
			char = (hexconv.Halfbyte[data[i+1]] << 4) | hexconv.Halfbyte[data[i+2]]
			buff = append(buff, char)
			i++
			offset = i + 2
			fallthrough // thereby checking whether the parsed char is printable
		default:
			if isProhibitedChar(char) {
				return into, buff, status.ErrBadEncoding
			}
		}
	}

	if escaped {
		buff = append(buff, data[offset:]...)
		key = uf.B2S(buff[buffptr:])
		escaped = false
	} else {
		key = uf.B2S(data[:])
	}

	if len(key) == 0 {
		return into, buff, status.ErrBadEncoding
	}

	return append(into, form.Data{Name: key}), buff, nil

parseValue:
	buffptr = len(buff)

	for i := 0; i < len(data); i++ {
		switch char := data[i]; char {
		case '%':
			escaped = true
			if i+2 >= len(data) {
				return into, buff, status.ErrBadEncoding
			}

			buff = append(buff, data[offset:i]...)
			a, b := hexconv.Halfbyte[data[i+1]], hexconv.Halfbyte[data[i+2]]
			if a|b == 0xFF {
				return into, buff, status.ErrBadEncoding
			}

			buff = append(buff, (a<<4)|b)
			i++
			offset = i + 2
		case '+':
			data[i] = ' '
		case '&':
			var value string

			if escaped {
				buff = append(buff, data[offset:i]...)
				value = uf.B2S(buff[buffptr:])
				escaped = false
			} else {
				value = uf.B2S(data[:i])
			}

			if len(key) == 0 {
				return into, buff, status.ErrBadEncoding
			}

			into = append(into, form.Data{
				Name:  key,
				Value: value,
			})

			data = data[i+1:]
			offset = 0
			goto parseKey
		}
	}

	var value string

	if escaped {
		buff = append(buff, data[offset:]...)
		value = uf.B2S(buff[buffptr:])
		escaped = false
	} else {
		value = uf.B2S(data)
	}

	if len(key) == 0 {
		return into, buff, status.ErrBadEncoding
	}

	into = append(into, form.Data{
		Name:  key,
		Value: value,
	})

	return into, buff, nil
}

func isProhibitedChar(c byte) bool {
	return c < 0x20 || c > 0x7e
}
