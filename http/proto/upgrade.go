package proto

import (
	"strings"

	"github.com/indigo-web/indigo/v2/internal/split"
)

func ChooseUpgrade(line string) Proto {
	// TODO: idk why but this variant is the fastest one, even comparing to
	//       the dumbest solution with inlined string iteration and hardcoded
	//       constant (comma-delimiter). Maybe, I am wrong? Hope there are
	//       better solutions
	iter := split.StringIter(line, ',')

	for {
		token, err := iter()
		if err != nil {
			return Unknown
		}

		if proto := parseUpgradeToken(strings.TrimSpace(token)); proto != Unknown {
			// pick the first supported protocol, as they are placed in an order of
			// preference
			return proto
		}
	}
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
