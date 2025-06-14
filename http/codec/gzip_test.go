package codec

import (
	"bytes"
	"fmt"
	"github.com/indigo-web/indigo/http"
	"github.com/indigo-web/indigo/transport/dummy"
	"github.com/klauspost/compress/gzip"
	"github.com/stretchr/testify/require"
	"io"
	"strconv"
	"strings"
	"testing"
)

func TestGZIP(t *testing.T) {
	dc := NewGZIP(16)
	require.NoError(t, dc.Reset(dummy.NewCircularClient(gzipped("Hello, world!")).OneTime()))
	decompressed, err := fetchAll(dc)
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

func fetchAll(source http.Fetcher) (string, error) {
	builder := strings.Builder{}

	for {
		data, err := source.Fetch()
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
