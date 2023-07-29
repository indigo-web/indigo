package strcomp

func EqualFold(a, b string) bool {
	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i]|0x20 != b[i]|0x20 {
			return false
		}
	}

	return true
}
