package strutil

import (
	"strings"

	"github.com/indigo-web/indigo/internal/hexconv"
)

// IsURLUnsafeChar tells whether it's safe to decode an urlencoded character.
func IsURLUnsafeChar(c byte) bool {
	return c == '/' || IsASCIINonprintable(c)
}

// URLDecode decodes an urlencoded string and tells whether the string was properly formed.
func URLDecode(str string) (string, bool) {
	var b strings.Builder
	b.Grow(len(str))
	s := str

	for len(s) > 0 {
		percent := strings.IndexByte(s, '%')
		if percent == -1 {
			break
		}

		b.WriteString(s[:percent])
		s = s[percent+1:]
		if len(s) < 2 {
			return "", false
		}

		c1, c2 := s[0], s[1]
		s = s[2:]
		x, y := hexconv.Halfbyte[c1], hexconv.Halfbyte[c2]
		if x|y == 0xFF {
			return "", false
		}

		char := (x << 4) | y
		if IsURLUnsafeChar(char) {
			b.Write([]byte{'%', c1 | 0x20, c2 | 0x20})
			continue
		}

		b.WriteByte(char)
	}

	b.WriteString(s)

	return b.String(), true
}
