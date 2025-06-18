package uri

// Normalize removes trailing slashes, as all request paths are also trimmed, resulting
// in consensus between these two.
func Normalize(path string) string {
	for i := len(path) - 1; i > 1; i-- {
		if path[i] != '/' {
			return path[:i+1]
		}
	}

	return path
}
