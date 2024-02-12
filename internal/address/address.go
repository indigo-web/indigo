package address

import (
	"net"
	"strings"
)

const DefaultAddr = "0.0.0.0"

func Format(addr string) string {
	if len(removePort(addr)) == 0 {
		// only port is presented
		return DefaultAddr + addr
	}

	return addr
}

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
