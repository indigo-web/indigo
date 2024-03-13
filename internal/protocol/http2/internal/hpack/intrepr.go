package hpack

func decode(n int, value []byte) (decoded uint64) {
	// first value is always the prefix
	// null first 8-n bytes, which don't relate to the prefix
	decoded = uint64(value[0] & (0xff >> (8 - n)))

	if decoded < (1<<n)-1 {
		return decoded
	}

	for i, b := range value[1:] {
		decoded += uint64(b&127) * (1 << (i * 7))
		if b&0x80 == 0 {
			break
		}
	}

	return decoded
}
