package http

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestResponse(t *testing.T) {
	t.Run("JSON", func(t *testing.T) {
		response := NewResponse()
		m := []int{1, 2, 3}
		resp, err := response.WithJSON(m)
		require.NoError(t, err)
		require.Equal(t, "[1,2,3]", string(resp.Body))
		require.Equal(t, "application/json", resp.ContentType)
	})
}
