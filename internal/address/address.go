package address

import (
	"net"
	"strings"
)

func IsLocalhost(addr string) bool {
	return strings.EqualFold(removePort(addr), "localhost")
}

func IsIP(addr string) bool {
	return net.ParseIP(removePort(addr)) != nil
}

func removePort(addr string) string {
	colon := strings.IndexByte(addr, ':')
	if colon != -1 {
		return addr[:colon]
	}

	return addr
}
