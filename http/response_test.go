package http

import (
	"testing"

	"github.com/indigo-web/indigo/kv"
	"github.com/stretchr/testify/require"
)

func TestResponse(t *testing.T) {
	t.Run("JSON", func(t *testing.T) {
		response := NewResponse()
		m := []int{1, 2, 3}
		resp, err := response.TryJSON(m)
		require.NoError(t, err)
		require.Equal(t, "[1,2,3]", string(resp.fields.Buffer))
		contentType := kv.NewFromPairs(resp.fields.Headers).Value("Content-Type")
		require.Equal(t, "application/json", contentType)
	})
}
