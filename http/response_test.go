package http

import (
	"errors"
	"github.com/indigo-web/indigo/http/status"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestResponse(t *testing.T) {
	t.Run("JSON", func(t *testing.T) {
		response := NewResponse()
		m := []int{1, 2, 3}
		resp, err := response.TryJSON(m)
		require.NoError(t, err)
		require.Equal(t, "[1,2,3]", string(resp.Reveal().Body))
		require.Equal(t, "application/json", resp.Reveal().ContentType)
	})
}

func BenchmarkResponse_WithError(b *testing.B) {
	resp := NewResponse()
	knownErr := status.ErrBadRequest
	unknownErr := errors.New("some crap happened, unable to recover")

	b.Run("KnownError", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			resp.Error(knownErr)
		}
	})

	b.Run("UnknownError", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			resp.Error(unknownErr)
		}
	})
}
