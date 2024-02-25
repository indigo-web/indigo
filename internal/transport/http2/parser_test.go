package http2

import (
	"github.com/indigo-web/indigo/internal/transport"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestParser(t *testing.T) {
	t.Run("preface and settings", func(t *testing.T) {
		raw := "PRI * HTTP/2.0\r\n\r\nSM\r\n\r\n\x00\x00\x18\x04\x00\x00\x00\x00\x00\x00\x01\x00\x01\x00\x00\x00" +
			"\x02\x00\x00\x00\x00\x00\x04\x00`\x00\x00\x00\x06\x00\x04\x00\x00\x00\x00\x04\b\x00\x00\x00\x00\x00" +
			"\x00\xef\x00\x01"
		parser := NewParser()
		state, extra, err := parser.Parse([]byte(raw))
		require.NoError(t, err)
		require.Empty(t, extra)
		require.Equal(t, transport.Pending, state)
	})
}
