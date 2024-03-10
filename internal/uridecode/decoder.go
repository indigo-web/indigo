package uridecode

import (
	"bytes"
	"github.com/indigo-web/indigo/http/status"
	"github.com/indigo-web/indigo/internal/hexconv"
)

// Decode normalizes the URI by translating escaped characters into their
// true form
func Decode(src, buff []byte) ([]byte, error) {
	for i := bytes.IndexByte(src, '%'); i != -1; i = bytes.IndexByte(src, '%') {
		if i >= len(src)-2 {
			return nil, status.ErrURIDecoding
		}

		buff = append(buff, src[:i]...)
		buff = append(buff, hexconv.Parse(src[i+1])<<4|hexconv.Parse(src[i+2]))
		src = src[i+3:]
	}

	if len(buff) == 0 {
		return src, nil
	}

	return append(buff, src...), nil
}
