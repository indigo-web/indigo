package cookie

import (
	"errors"
	"github.com/indigo-web/indigo/internal/keyvalue"
	"strings"
)

// Jar is a key-value storage for cookies. Key-value pairs consists of strings,
// not cookie.Cookie, as it would lead to space wasting and require a separate
// data structure
type Jar = *keyvalue.Storage

func NewJar() Jar {
	return keyvalue.New()
}

func NewJarPreAlloc(n int) Jar {
	return keyvalue.NewPreAlloc(n)
}

var ErrBadCookie = errors.New("cookie has a malformed syntax")

// Parse parses cookies, received from a user-agent. These are basically key-value pairs,
// so the function isn't applicable for Set-Cookie values
func Parse(jar Jar, data string) (err error) {
	for len(data) > 0 {
		eq := strings.IndexByte(data, '=')
		if eq == -1 {
			break
		}

		key := data[:eq]
		data = data[eq+1:]

		if len(key) == 0 {
			return ErrBadCookie
		}

		var value string

		if cs := strings.IndexByte(data, ';'); cs != -1 {
			value, data = data[:cs], stripSpace(data[cs+1:])
		} else {
			value, data = data, ""
		}

		// empty value is fine (probably, I have no idea if it's so)
		jar.Add(key, value)
	}

	if len(data) != 0 {
		return ErrBadCookie
	}

	return nil
}

func stripSpace(str string) string {
	if len(str) > 0 && str[0] == ' ' {
		return str[1:]
	}

	return str
}
