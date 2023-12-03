package address

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

// DefaultHost used in cases, when no host was defined. For example, the string ":8080"
const DefaultHost = "0.0.0.0"

type Address struct {
	Host string
	Port uint16
}

func Parse(addr string) (address Address, err error) {
	colon := strings.IndexByte(addr, ':')
	if colon == -1 {
		return address, fmt.Errorf("no port given")
	}

	host, rawPort := addr[:colon], addr[colon+1:]
	if len(host) == 0 {
		host = DefaultHost
	}
	if len(rawPort) == 0 {
		return address, fmt.Errorf("port cannot be empty")
	}

	port, err := strconv.Atoi(rawPort)
	if err != nil || port < 0 || port > math.MaxUint16 {
		return address, fmt.Errorf("invalid port: %s", rawPort)
	}

	return Address{
		Host: host,
		Port: uint16(port),
	}, nil
}

func (a Address) SetPort(newPort uint16) Address {
	a.Port = newPort

	return a
}

func (a Address) String() string {
	return fmt.Sprintf("%s:%d", a.Host, a.Port)
}
