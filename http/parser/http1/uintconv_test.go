package http1

import (
	"github.com/fakefloordiv/indigo/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseUIntValidCases(t *testing.T) {
	t.Run("NilSlice", func(t *testing.T) {
		num, err := parseUint("")
		require.Nil(t, err)
		require.Equal(t, uint(0), num)
	})

	t.Run("SingleNum", func(t *testing.T) {
		num, err := parseUint("1")
		require.Nil(t, err)
		require.Equal(t, uint(1), num)
	})

	t.Run("TenNums", func(t *testing.T) {
		num, err := parseUint("1234567890")
		require.Nil(t, err)
		require.Equal(t, uint(1234567890), num)
	})

	t.Run("LeadingZero", func(t *testing.T) {
		num, err := parseUint("0042")
		require.Nil(t, err)
		require.Equal(t, uint(42), num)
	})
}

func TestParseUIntInvalidCases(t *testing.T) {
	t.Run("InvalidSingleChar", func(t *testing.T) {
		num, err := parseUint("123g456")
		require.Equal(t, uint(0), num)
		require.Equal(t, err, http.ErrBadRequest)
	})

	t.Run("InvalidWholeNumber", func(t *testing.T) {
		num, err := parseUint("hello, world!")
		require.Equal(t, uint(0), num)
		require.Equal(t, err, http.ErrBadRequest)
	})
}
