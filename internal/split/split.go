package split

import "io"

// Iterator returns new string (or error) on every next call
type Iterator func() (string, error)

// StringIter returns Iterator that walks by a string
func StringIter(str string, sep byte) Iterator {
	var offset int

	return func() (string, error) {
		if len(str) == 0 {
			return "", io.EOF
		}

		for i := offset; i < len(str); i++ {
			if str[i] == sep {
				piece := str[offset:i]
				offset = i + 1

				return piece, nil
			}
		}

		piece := str[offset:]
		str = ""

		return piece, nil
	}
}
