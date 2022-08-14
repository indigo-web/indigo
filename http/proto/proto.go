package proto

type Proto uint8

const (
	Unknown Proto = 0
	HTTP09  Proto = iota + 9 // HTTP09 == 9, HTTP10 = 10, HTTP11 = 11
	HTTP10
	HTTP11
	// HTTP2, HTTP3 - will be added when implemented
)

var (
	http09 = []byte("HTTP/0.9")
	http10 = []byte("HTTP/1.0")
	http11 = []byte("HTTP/1.1")
)

func Parse(proto string) Proto {
	switch proto {
	case "HTTP/0.9":
		return HTTP09
	case "HTTP/1.0":
		return HTTP10
	case "HTTP/1.1":
		return HTTP11
	default:
		return Unknown
	}
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
