package proto

import (
	"github.com/indigo-web/indigo/internal/cutbyte"
	"strings"
)

func ChooseUpgrade(line string) Proto {
	for len(line) > 0 {
		var token string
		token, line = cutbyte.Cut(line, ',')

		if proto := parseUpgradeToken(strings.TrimSpace(token)); proto != Unknown {
			// pick the first supported protocol, as they are placed in an order of
			// preference
			return proto
		}
	}

	return Unknown
}

// parseUpgradeToken simply parses an upgrade-token to the respective protocol enum
func parseUpgradeToken(token string) Proto {
	switch token {
	case "http/0.9", "HTTP/0.9":
		return HTTP09
	case "http/1.0", "HTTP/1.0":
		return HTTP10
	case "http/1.1", "HTTP/1.1":
		return HTTP11
	case "h2c", "H2C":
		return HTTP2
	case "websocket", "WebSocket", "WEBSOCKET":
		return WebSocket
	}

	return Unknown
}
