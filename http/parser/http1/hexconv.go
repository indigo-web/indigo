package http1

func isHex(char byte) bool {
	switch {
	case '0' <= char && char <= '9':
		return true
	case 'a' <= char && char <= 'f':
		return true
	case 'A' <= char && char <= 'F':
		return true
	}
	return false
}

func unHex(char byte) byte {
	switch {
	case '0' <= char && char <= '9':
		return char - '0'
	case 'a' <= char && char <= 'f':
		return char - 'a' + 10
	case 'A' <= char && char <= 'F':
		return char - 'A' + 10
	}
	return 0
}
