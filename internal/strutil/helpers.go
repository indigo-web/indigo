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

// ParseQualifier returns an int in range [0, 10] representing the qualifier value.
// All values below the 0.1 resolution are ignored. Invalid values result in 0. But
// keep in mind that 0 is also a valid value.
func ParseQualifier(q string) int {
	const sampleQualifier = "q=p.q"

	if len(q) < len(sampleQualifier) {
		return 0
	}

	// ignore all values below the 0.1 resolution
	qualifier := ctoi(q[2])*10 + ctoi(q[4])
	if qualifier < 0 || qualifier > 10 {
		qualifier = 0
	}

	return qualifier
}

func ctoi(char byte) int {
	return int(char - '0')
}

func Unquote(str string) string {
	if len(str) > 1 && str[0] == '"' && str[len(str)-1] == '"' {
		return str[1 : len(str)-1]
	}

	return str
}
