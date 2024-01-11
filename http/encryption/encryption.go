package encryption

type Encryption uint8

const (
	Plain Encryption = iota
	TLS
	// Extend is used to extend the enums, if custom encryption is used
	Extend
)
