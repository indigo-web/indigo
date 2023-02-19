package proto

type Proto uint8

const (
	Unknown Proto = 0
	HTTP09  Proto = 1 << iota
	HTTP10
	HTTP11
	HTTP2

	WebSocket

	HTTP1 = HTTP09 | HTTP10 | HTTP11
)

var (
	http09 = []byte("HTTP/0.9 ")
	http10 = []byte("HTTP/1.0 ")
	http11 = []byte("HTTP/1.1 ")
)

func Parse(major, minor uint8) Proto {
	switch major {
	case 0:
		switch minor {
		case 9:
			return HTTP09
		}
	case 1:
		switch minor {
		case 0:
			return HTTP10
		case 1:
			return HTTP11
		}
	}

	return Unknown
}

func ToBytes(proto Proto) []byte {
	switch proto {
	case HTTP09:
		return http09
	case HTTP10:
		return http10
	case HTTP11:
		return http11
	default:
		return nil
	}
}
