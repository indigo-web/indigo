package decode

type GZIPDecoder struct {
}

func NewGZIPDecoder() DecoderFactory {
	return &GZIPDecoder{}
}

func (g *GZIPDecoder) New() DecoderFunc {
	return g.decode
}

func (g *GZIPDecoder) decode(input []byte) (output []byte, err error) {
	panic("gzip decoder not implemented yet")
}
