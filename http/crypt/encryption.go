package crypt

type Encryption uint8

const (
	Plain Encryption = 0
	SSL   Encryption = 1 << (iota - 1)
	TLSv10
	TLSv11
	TLSv12
	TLSv13
	Unknown
)

func (e Encryption) IsTLS() bool {
	return e&(SSL|TLSv10|TLSv11|TLSv12|TLSv13) != 0
}

func (e Encryption) IsSafe() bool {
	return e != Plain
}
