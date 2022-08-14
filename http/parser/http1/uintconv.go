package http1

import (
	"indigo/errors"
)

/*
parseUint is a tiny implementation of strconv.Atoi, but using directly bytes array,
and returning only one error in case of shit - InvalidContentLength
Parses 10-numeral system integers
*/
func parseUint(raw []byte) (num int, err error) {
	for _, char := range raw {
		char -= '0'

		if char > 9 {
			return 0, errors.ErrBadRequest
		}

		num = num*10 + int(char)
	}

	return num, nil
}
