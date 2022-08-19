package http1

import (
	"indigo/errors"
)

// parseUint is a tiny implementation of strconv.Atoi, but reading
// from byte slice instead of string (also it's a bit faster (I hope))
func parseUint(raw []byte) (num uint, err error) {
	for _, char := range raw {
		char -= '0'

		if char > 9 {
			return 0, errors.ErrBadRequest
		}

		num = num*10 + uint(char)
	}

	return num, nil
}
