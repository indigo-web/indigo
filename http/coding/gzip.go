package coding

import (
	"bytes"
	"compress/gzip"
	"io"
)

type GZIP struct {
	inBuff  *bytes.Buffer
	outBuff []byte
	reader  *gzip.Reader
	writer  *gzip.Writer
}

func NewGZIP(outBuff []byte) Coding {
	buff := bytes.NewBuffer(nil)
	// gzip.NewWriterLevel never returns an error, so it's safe to just ignore it
	writer, _ := gzip.NewWriterLevel(nil, gzip.DefaultCompression)

	return &GZIP{
		inBuff:  buff,
		outBuff: outBuff,
		reader:  new(gzip.Reader),
		writer:  writer,
	}
}

func (g *GZIP) Token() string {
	return "gzip"
}

func (g *GZIP) Decode(input []byte) (output []byte, err error) {
	g.inBuff.Reset()
	g.inBuff.Write(input)
	if err = g.reader.Reset(g.inBuff); err != nil {
		return nil, err
	}

	n, err := g.reader.Read(g.outBuff)
	switch err {
	case nil, io.EOF:
		return g.outBuff[:n], err
	default:
		return nil, err
	}
}

func (g *GZIP) Encode(input []byte) (output []byte, err error) {
	g.inBuff.Reset()
	g.writer.Reset(g.inBuff)
	_, err = g.writer.Write(input)
	if err != nil {
		return nil, err
	}

	err = g.writer.Flush()
	return g.inBuff.Bytes(), err
}
