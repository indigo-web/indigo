package strutil

import (
	"iter"
)

// a-z A-Z 0-9 ()[]{}-_<>.,/|%"
// % is included, as WalkKV does not decode key or value, therefore urlencoded values must
// not appear as unsafe characters
var safeChars = [256]bool{
	false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false,
	false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false,
	false, false, true, false, false, true, false, false, true, true, false, true, true, true, true, true,
	true, true, true, true, true, true, true, true, true, true, false, false, true, false, true, false,
	false, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true,
	true, true, true, true, true, true, true, true, true, true, true, true, false, true, false, true,
	false, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true,
	true, true, true, true, true, true, true, true, true, true, true, true, true, true, false, false,
	false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false,
	false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false,
	false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false,
	false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false,
	false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false,
	false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false,
	false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false,
	false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false,
}

func WalkKV(data string) iter.Seq2[string, string] {
	return func(yield func(string, string) bool) {
		var key string

	paramKey:
		for i := 0; i < len(data); i++ {
			c := data[i]

			if c == '=' {
				key = data[:i]
				data = data[i+1:]
				goto paramValue
			}

			if !safeChars[c] {
				yield("", "")
				return
			}
		}

		yield(data, "")
		return

	paramValue:
		for i := 0; i < len(data); i++ {
			c := data[i]

			if c == ';' {
				value := data[:i]
				data = LStripWS(data[i+1:])

				if !yield(key, Unquote(value)) {
					return
				}

				goto paramKey
			}

			if !safeChars[c] {
				yield("", "")
				return
			}
		}

		yield(key, Unquote(data))
		return
	}
}
