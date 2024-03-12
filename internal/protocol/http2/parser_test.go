package http2

import (
	"github.com/indigo-web/indigo/internal/tcp/dummy"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestParser(t *testing.T) {
	t.Run("readN", func(t *testing.T) {
		client := dummy.NewCircularClient([]byte("Hello, world!"))
		parser := NewParser(client)
		data, err := parser.readN(1)
		require.NoError(t, err)
		require.Equal(t, "H", string(data))
		data, err = parser.readN(5)
		require.NoError(t, err)
		require.Equal(t, "ello,", string(data))
	})

	t.Run("preface and settings", func(t *testing.T) {
		//raw := "PRI * HTTP/2.0\r\n\r\nSM\r\n\r\n\x00\x00\x18\x04\x00\x00\x00\x00\x00\x00\x01\x00\x01\x00\x00\x00" +
		//	"\x02\x00\x00\x00\x00\x00\x04\x00`\x00\x00\x00\x06\x00\x04\x00\x00\x00\x00\x04\b\x00\x00\x00\x00\x00" +
		//	"\x00\xef\x00\x01"
		//parser := NewParser()
		//frame, err := parser.Parse()
		//require.NoError(t, err)
	})
}
