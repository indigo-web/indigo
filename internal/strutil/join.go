package strutil

import (
	"iter"
	"strings"
)

// Join works in the same way as the strings.Join does, except that it operates an iterator
// as opposed to greedy string slice.
func Join(elems iter.Seq[string], sep string) string {
	var b strings.Builder

	for elem := range elems {
		if b.Len() > 0 {
			b.WriteString(sep)
		}

		b.WriteString(elem)
	}

	return b.String()
}
