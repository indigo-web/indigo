package proto

import (
	"strings"
)

func ChooseUpgrade(line string) Protocol {
	for len(line) > 0 {
		var token string
		token, line = cutbyte(line, ',')

		if proto := parseUpgradeToken(strings.TrimSpace(token)); proto != Unknown {
			// pick the first supported protocol, as they are placed in an order of
			// preference
			return proto
		}
	}

	return Unknown
}

// parseUpgradeToken simply parses an upgrade-token to the respective protocol enum
func parseUpgradeToken(token string) Protocol {
	switch token {
	case "http/1.0", "HTTP/1.0":
		return HTTP10
	case "http/1.1", "HTTP/1.1":
		return HTTP11
	case "h2c", "H2C":
		return HTTP2
	}

	return Unknown
}

func cutbyte(str string, sep byte) (prefix, postfix string) {
	for i := 0; i < len(str); i++ {
		if str[i] == sep {
			return str[:i], str[i+1:]
		}
	}

	return str, ""
}
