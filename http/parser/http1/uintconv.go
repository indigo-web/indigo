package http1

import (
	"github.com/fakefloordiv/indigo/http"
)

// parseUint is a tiny implementation of strconv.Atoi, but reading
// from byte slice instead of string (also it's a bit faster (I hope))
func parseUint(raw []byte) (num uint, err error) {
	for _, char := range raw {
		char -= '0'

		if char > 9 {
			return 0, http.ErrBadRequest
		}

		num = num*10 + uint(char)
	}

	return num, nil
}
