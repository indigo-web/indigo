package urlencoded

import (
	"bytes"
	"github.com/indigo-web/indigo/http/status"
	"github.com/indigo-web/indigo/internal/hexconv"
	"github.com/indigo-web/utils/uf"
)

// Decode decodes data into the given buffer, but omits it if there's no data to be
// decoded. `into` can be data[:0] as well in order to decode "into itself".
func Decode(src, dst []byte) (decoded, buffer []byte, err error) {
	percent := bytes.IndexByte(src, '%')
	if percent == -1 {
		return src, dst, nil
	}

	for percent != -1 {
		if percent >= len(src)-2 {
			return nil, dst, status.ErrURLDecoding
		}

		dst = append(dst, src[:percent]...)
		a, b := hexconv.Halfbyte[src[percent+1]], hexconv.Halfbyte[src[percent+2]]
		if a|b > 0x0f {
			return nil, dst, status.ErrURLDecoding
		}

		dst = append(dst, (a<<4)|b)
		src = src[percent+3:]
		percent = bytes.IndexByte(src, '%')
	}

	dst = append(dst, src...)
	return dst, dst, nil
}

// ExtendedDecode is the same as Decode, but on top also decodes + as spaces.
func ExtendedDecode(src, dst []byte) (decoded, buffer []byte, err error) {
	dsthead := len(dst)
	modified := false

loop:
	for i, c := range src {
		if c == '+' {
			modified = true
			dst = append(dst, src[:i]...)
			dst = append(dst, ' ')
			src = src[i+1:]
			goto loop
		} else if c == '%' {
			modified = true

			if len(src)-i < 3 {
				return nil, dst, status.ErrURLDecoding
			}

			a, b := hexconv.Halfbyte[src[i+1]], hexconv.Halfbyte[src[i+2]]
			if a|b > 0x0f {
				return nil, dst, status.ErrURLDecoding
			}
			dst = append(dst, src[:i]...)
			dst = append(dst, (a<<4)|b)
			src = src[i+3:]
			goto loop
		}
	}

	if !modified {
		return src, dst, nil
	}

	dst = append(dst, src...)
	return dst[dsthead:], dst, nil
}

func ExtendedDecodeString(src string, buff []byte) (decoded string, buffer []byte, err error) {
	d, buffer, err := ExtendedDecode(uf.S2B(src), buff)
	return uf.B2S(d), buffer, err
}
