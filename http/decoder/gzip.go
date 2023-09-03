package decoder

import (
	"bytes"
	"compress/gzip"
	"io"
)

type GZIPDecoder struct {
	inputBuff  *bytes.Buffer
	outputBuff []byte
	reader     *gzip.Reader
}

func NewGZIPDecoder(outputBuff []byte) Decoder {
	buff := bytes.NewBuffer(nil)

	return &GZIPDecoder{
		inputBuff:  buff,
		outputBuff: outputBuff,
	}
}

func (g *GZIPDecoder) Decode(input []byte) (output []byte, err error) {
	g.inputBuff.Reset()
	g.inputBuff.Write(input)
	reader, err := gzip.NewReader(g.inputBuff)
	if err != nil {
		return nil, err
	}

	n, err := reader.Read(g.outputBuff)
	switch err {
	case nil, io.EOF:
		return g.outputBuff[:n], nil
	default:
		return nil, err
	}
}
