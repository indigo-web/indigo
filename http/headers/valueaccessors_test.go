package headers

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetParam(t *testing.T) {
	value := "hello;world=true;another=earth"

	t.Run("Positive_World", func(t *testing.T) {
		substr := ";world="
		require.Equal(t, "true", getParam(value, substr, ""))
	})

	t.Run("Positive_Another", func(t *testing.T) {
		substr := ";another="
		require.Equal(t, "earth", getParam(value, substr, ""))
	})

	t.Run("Negative", func(t *testing.T) {
		substr := ";unknown="
		require.Empty(t, getParam(value, substr, ""))
	})
}

func TestQualityOf(t *testing.T) {
	value1 := "text/html;q=0.5;charset=utf8"
	value2 := "text/html;charset=utf8;q=0.5"
	valueNoQ := "text/html;charset=utf8"
	valueInvalidQ := "text/html;q=0.a"

	t.Run("Positive_1", func(t *testing.T) {
		require.Equal(t, 5, QualityOf(value1))
	})

	t.Run("Positive_2", func(t *testing.T) {
		require.Equal(t, 5, QualityOf(value2))
	})

	t.Run("Negative_NoQ", func(t *testing.T) {
		require.Equal(t, 9, QualityOf(valueNoQ))
	})

	t.Run("Negative_InvalidQ", func(t *testing.T) {
		require.Equal(t, 9, QualityOf(valueInvalidQ))
	})
}

func TestValueOf(t *testing.T) {
	valueWithoutParams := "text/html"
	valueWithParams := "text/html;q=0.9"

	t.Run("WithoutParams", func(t *testing.T) {
		require.Equal(t, valueWithoutParams, ValueOf(valueWithoutParams))
	})

	t.Run("WithParams", func(t *testing.T) {
		require.Equal(t, "text/html", ValueOf(valueWithParams))
	})
}
