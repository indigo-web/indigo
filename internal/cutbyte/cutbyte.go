package cutbyte

func Cut(str string, sep byte) (prefix, postfix string) {
	for i := 0; i < len(str); i++ {
		if str[i] == sep {
			return str[:i], str[i+1:]
		}
	}

	return str, ""
}
