package address

import (
	"net"
	"strings"
)

const DefaultAddr = "0.0.0.0"

func Normalize(addr string) string {
	if len(stripPort(addr)) == 0 {
		// only port is presented
		return DefaultAddr + addr
	}

	return addr
}

func IsLocalhost(addr string) bool {
	return strings.EqualFold(stripPort(addr), "localhost")
}

func IsIP(addr string) bool {
	return net.ParseIP(stripPort(addr)) != nil
}

func stripPort(addr string) string {
	colon := strings.IndexByte(addr, ':')
	if colon != -1 {
		return addr[:colon]
	}

	return addr
}
