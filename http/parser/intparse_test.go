package parser

import (
	"indigo/errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseUIntValidCases(t *testing.T) {
	t.Run("NilSlice", func(t *testing.T) {
		num, err := parseUint(nil)
		require.Nil(t, err)
		require.Equal(t, 0, num)
	})

	t.Run("SingleNum", func(t *testing.T) {
		num, err := parseUint([]byte("1"))
		require.Nil(t, err)
		require.Equal(t, 1, num)
	})

	t.Run("TenNums", func(t *testing.T) {
		num, err := parseUint([]byte("1234567890"))
		require.Nil(t, err)
		require.Equal(t, 1234567890, num)
	})

	t.Run("LeadingZero", func(t *testing.T) {
		num, err := parseUint([]byte("0042"))
		require.Nil(t, err)
		require.Equal(t, 42, num)
	})
}

func TestParseUIntInvalidCases(t *testing.T) {
	t.Run("InvalidSingleChar", func(t *testing.T) {
		num, err := parseUint([]byte("123g456"))
		require.Equal(t, 0, num)
		require.Equal(t, err, errors.ErrInvalidContentLength)
	})

	t.Run("InvalidWholeNumber", func(t *testing.T) {
		num, err := parseUint([]byte("hello, world!"))
		require.Equal(t, 0, num)
		require.Equal(t, err, errors.ErrInvalidContentLength)
	})
}