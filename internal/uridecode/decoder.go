package uridecode

import (
	"bytes"
	"github.com/indigo-web/indigo/http/status"
	"github.com/indigo-web/indigo/internal/hex"
)

func Decode(src, buff []byte) ([]byte, error) {
	for {
		separator := bytes.IndexByte(src, '%')
		if separator == -1 {
			if len(buff) == 0 {
				return src, nil
			}

			return append(buff, src...), nil
		}

		if len(src[separator+1:]) < 2 || !hex.Is(src[separator+1]) || !hex.Is(src[separator+2]) {
			return nil, status.ErrURIDecoding
		}

		buff = append(buff, src[:separator]...)
		buff = append(buff, (hex.Un(src[separator+1])<<4)|hex.Un(src[separator+2]))
		src = src[separator+3:]
	}
}
