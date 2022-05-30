package internal

/*
toLowercase does stuff directly on the array, without allocating a new buffer,
so original buffer will be affected
*/
func toLowercase(data []byte) {
	for i, char := range data {
		data[i] = char | 0x20
	}
}
