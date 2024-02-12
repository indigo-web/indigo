package encryption

type Token uint8

const (
	Plain Token = iota
	TLS
	// Extend is used to extend the enums, if custom encryption is used
	Extend
)
