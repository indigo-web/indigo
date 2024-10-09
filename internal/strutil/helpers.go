package strutil

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

func Unquote(str string) string {
	if len(str) > 1 && str[0] == '"' && str[len(str)-1] == '"' {
		return str[1 : len(str)-1]
	}

	return str
}
