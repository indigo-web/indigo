package strutil

const (
	defaultAddress = "0.0.0.0"
	defaultPort    = "0"
)

func NormalizeAddress(addr string) string {
	if len(addr) == 0 {
		// the function should never receive empty address anyway
		return addr
	}

	if addr[0] == ':' {
		addr = defaultAddress + addr
	}

	return addr
}
