package strutil

import "strings"

func LStripWS(str string) string {
	for i, c := range str {
		switch c {
		// TODO: consider adding more whitespace characters?
		case ' ', '\t':
		default:
			return str[i:]
		}
	}

	return ""
}

func RStripWS(str string) string {
	for i := len(str); i > 0; i-- {
		switch str[i-1] {
		case ' ', '\t':
		default:
			return str[:i]
		}
	}

	return ""
}

// CutParams behaves exactly as strings.Cut, but strips whitespaces between value
// and the first-encountered parameter in addition.
func CutParams(header string) (params string) {
	_, params = CutHeader(header)
	return params
}

func CutHeader(header string) (value, params string) {
	sep := strings.IndexByte(header, ';')
	if sep == -1 {
		return header, ""
	}

	return header[:sep], LStripWS(header[sep+1:])
}

func Unquote(str string) string {
	if len(str) > 1 && str[0] == '"' && str[len(str)-1] == '"' {
		return str[1 : len(str)-1]
	}

	return str
}
