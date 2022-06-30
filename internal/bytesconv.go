package internal

/*
ToLowercase applies on source data. So yes, it's dirty function, but does its stuff
blazingly fast
*/
func ToLowercase(data []byte) {
	for i, char := range data {
		data[i] = char | 0x20
	}
}
