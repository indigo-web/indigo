package hexconv

var decodeTable = [256]byte{
	'0': 0x0,
	'1': 0x1,
	'2': 0x2,
	'3': 0x3,
	'4': 0x4,
	'5': 0x5,
	'6': 0x6,
	'7': 0x7,
	'8': 0x8,
	'9': 0x9,
	'a': 0xa,
	'b': 0xb,
	'c': 0xc,
	'd': 0xd,
	'e': 0xe,
	'f': 0xf,
	'A': 0xA,
	'B': 0xB,
	'C': 0xC,
	'D': 0xD,
	'E': 0xE,
	'F': 0xF,
}

// Parse returns char value + 1 IF char is a valid hex, 0 otherwise.
// So in order to treat the returned value correctly, check whether it's 0
func Parse(char byte) byte {
	return decodeTable[char]
}
