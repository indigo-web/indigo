package domain

import "strings"

func Normalize(domain string) string {
	if strings.HasPrefix(domain, "www.") {
		domain = domain[len("www."):]
	}

	for i := len(domain); i >= 0; i-- {
		if domain[i] == '.' {
			break
		} else if domain[i] == ':' {
			switch port := domain[i+1:]; port {
			case "80", "443":
				// trim only default ports. Non-default must always be presented
				domain = domain[:len(port)-1]
			}

			break
		}
	}

	return domain
}

func TrimPort(domain string) string {
	if colon := strings.IndexByte(domain, ':'); colon != -1 {
		return domain[:colon]
	}

	return domain
}
