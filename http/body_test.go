package http

import (
	"io"
	"testing"

	"github.com/indigo-web/indigo/config"
	"github.com/indigo-web/indigo/transport/dummy"
	"github.com/stretchr/testify/require"
)

func TestBody(t *testing.T) {
	t.Run("reader", func(t *testing.T) {
		data := dummy.NewMockClient([]byte("Hello, world!"))
		request := &Request{cfg: config.Default()}
		b := NewBody(data)
		b.Reset(request)

		buff := make([]byte, 12)
		n, err := b.Read(buff)
		require.NoError(t, err)
		require.Equal(t, "Hello, world", string(buff[:n]))

		b.Reset(request)
		n, err = b.Read(buff)
		require.Empty(t, string(buff[:n]))
		require.EqualError(t, err, io.EOF.Error())
	})
}
