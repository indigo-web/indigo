package decode

type GZIPDecompressor struct {
}

func (g *GZIPDecompressor) Decompress(input []byte) (output []byte, err error) {
	panic("gzip decoder not implemented yet")
}
