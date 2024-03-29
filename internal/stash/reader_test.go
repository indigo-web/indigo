package stash

import (
	"github.com/stretchr/testify/require"
	"io"
	"testing"
)

type retriever struct {
	data string
}

func (r retriever) Retrieve() ([]byte, error) {
	return []byte(r.data), io.EOF
}

func TestReader(t *testing.T) {
	t.Run("both data and error simultaneously", func(t *testing.T) {
		r := New(retriever{"hello, world"}.Retrieve)
		buff := make([]byte, 64)
		n, err := r.Read(buff)
		require.Equal(t, 12, n)
		require.Equal(t, "hello, world", string(buff[:n]))
		require.EqualError(t, err, io.EOF.Error())
	})

	t.Run("multiple reads", func(t *testing.T) {
		r := New(retriever{"hello, world"}.Retrieve)
		buff := make([]byte, 2)
		data, err := readfull(r, buff)
		require.NoError(t, err)
		require.Equal(t, "hello, world", string(data))
	})
}

func readfull(from io.Reader, buff []byte) ([]byte, error) {
	var full []byte

	for {
		n, err := from.Read(buff)
		full = append(full, buff[:n]...)
		switch err {
		case nil:
		case io.EOF:
			return full, nil
		default:
			return full, err
		}
	}
}
