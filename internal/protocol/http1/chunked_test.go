package http1

import (
	"github.com/indigo-web/indigo/http/status"
	"github.com/stretchr/testify/require"
	"io"
	"testing"
)

func feed(c *chunkedParser, input []byte) (output, extra []byte, err error) {
	for len(input) > 0 {
		var data []byte
		data, input, err = c.Parse(input)
		output = append(output, data...)
		switch err {
		case nil:
		case io.EOF:
			return output, input, nil
		default:
			return output, input, err
		}
	}

	return output, nil, nil
}

func TestChunked(t *testing.T) {
	t.Run("just trailer", func(t *testing.T) {
		p := newChunkedParser()
		output, extra, err := feed(&p, []byte("0\r\n\r\n"))
		require.NoError(t, err)
		require.Empty(t, extra)
		require.Empty(t, output)
	})

	t.Run("trailer with field lines", func(t *testing.T) {
		p := newChunkedParser()
		output, extra, err := feed(&p, []byte("0\r\nHello: world\r\nworld: Hello\r\n\r\n"))
		require.NoError(t, err)
		require.Empty(t, extra)
		require.Empty(t, output)
	})

	testSimpleChunked := func(t *testing.T, p *chunkedParser) {
		output, extra, err := feed(p, []byte("d\r\nHello, world!\r\n0\r\n\r\n"))
		require.NoError(t, err)
		require.Empty(t, extra)
		require.Equal(t, "Hello, world!", string(output))
	}

	t.Run("single simple small chunk", func(t *testing.T) {
		p := newChunkedParser()
		testSimpleChunked(t, &p)
	})

	t.Run("reusability", func(t *testing.T) {
		p := newChunkedParser()

		for range 10 {
			testSimpleChunked(t, &p)
		}
	})

	t.Run("extension", func(t *testing.T) {
		p := newChunkedParser()
		output, extra, err := feed(&p, []byte("d;hello=world\r\nHello, world!\r\n0; checksum=no one cares\r\n\r\n"))
		require.NoError(t, err)
		require.Empty(t, extra)
		require.Equal(t, "Hello, world!", string(output))
	})

	t.Run("LF use", func(t *testing.T) {
		p := newChunkedParser()
		output, extra, err := feed(&p, []byte("d;hello=world\nHello, world!\n0; checksum=no one cares\n\n"))
		require.NoError(t, err)
		require.Empty(t, extra)
		require.Equal(t, "Hello, world!", string(output))
	})

	t.Run("fuzz input chunk sizes", func(t *testing.T) {
		sample := []byte("d;hello=world\r\nHello, world!\r\nd\r\nHello, Pavlo!\r\n0; checksum=no one cares\r\n\r\n")
		for i := range len(sample) - 1 {
			p := newChunkedParser()
			var output []byte

			for _, chunk := range scatter(sample, i+1) {
				out, extra, err := feed(&p, chunk)
				require.NoError(t, err)
				require.Empty(t, extra)
				output = append(output, out...)
			}

			require.Equal(t, "Hello, world!Hello, Pavlo!", string(output))
		}
	})

	t.Run("multiple hex characters", func(t *testing.T) {
		p := newChunkedParser()
		output, extra, err := feed(&p, []byte(
			"0000d\r\nHello, world!\r\n0000d\r\nHello, Pavlo!\r\n0\r\n\r\n",
		))
		require.NoError(t, err)
		require.Empty(t, extra)
		require.Equal(t, "Hello, world!Hello, Pavlo!", string(output))
	})

	t.Run("bad hex character", func(t *testing.T) {
		p := newChunkedParser()
		_, _, err := feed(&p, []byte("dg\r\nHello, world!\r\n0\r\n\r\n"))
		require.EqualError(t, err, status.ErrBadChunk.Error())
	})

	t.Run("too many length characters", func(t *testing.T) {
		p := newChunkedParser()
		_, _, err := feed(&p, []byte("00000000d\r\nHello, world!\r\n0\r\n\r\n"))
		require.EqualError(t, err, status.ErrBadChunk.Error())
	})
}
