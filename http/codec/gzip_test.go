package codec

import (
	"bytes"
	"fmt"
	"github.com/klauspost/compress/gzip"
	"github.com/stretchr/testify/require"
	"io"
	"strconv"
	"strings"
	"testing"
)

type dummyRetriever struct {
	content []byte
}

func (d *dummyRetriever) Retrieve() ([]byte, error) {
	return d.content, io.EOF
}

func TestGZIP(t *testing.T) {
	dc := NewGZIP(16)
	require.NoError(t, dc.Reset(&dummyRetriever{gzipped("Hello, world!")}))
	decompressed, err := retrieveAll(dc)
	fmt.Println("err:", err, "result:", strconv.Quote(decompressed))
}

func gzipped(text string) []byte {
	buff := bytes.NewBuffer(nil)
	c := gzip.NewWriter(buff)
	_, err := c.Write([]byte(text))
	if err != nil {
		panic("unexpected error during gzipping")
	}
	if c.Close() != nil {
		panic("unexpected error during closing gzip writer")
	}

	return buff.Bytes()
}

func retrieveAll(source Retriever) (string, error) {
	builder := strings.Builder{}

	for {
		data, err := source.Retrieve()
		builder.Write(data)
		switch err {
		case nil:
		case io.EOF:
			return builder.String(), nil
		default:
			return "", err
		}
	}
}
