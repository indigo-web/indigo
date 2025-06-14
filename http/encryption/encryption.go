package encryption

type Token uint8

const (
	Plain Token = 0
	SSL   Token = 1 << (iota - 1)
	TLSv10
	TLSv11
	TLSv12
	TLSv13
	Unknown
)

func (t Token) IsTLS() bool {
	return t&(SSL|TLSv10|TLSv11|TLSv12|TLSv13) != 0
}

func (t Token) IsSafe() bool {
	return t != Plain
}
