package uridecode

import (
	"github.com/indigo-web/indigo/http/status"
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
)

func TestDecode(t *testing.T) {
	t.Run("no escaping", func(t *testing.T) {
		str := "/hello"
		decoded, err := Decode([]byte(str), nil)
		require.NoError(t, err)
		require.Equal(t, "/hello", string(decoded))
	})

	t.Run("corners", func(t *testing.T) {
		str := "%2fhello%2f"
		decoded, err := Decode([]byte(str), nil)
		require.NoError(t, err)
		require.Equal(t, "/hello/", string(decoded))
	})

	t.Run("multiple consecutive", func(t *testing.T) {
		str := "%2f%20hello"
		decoded, err := Decode([]byte(str), nil)
		require.NoError(t, err)
		require.Equal(t, "/ hello", string(decoded))
	})

	t.Run("incomplete sequence", func(t *testing.T) {
		str := "%2"
		_, err := Decode([]byte(str), nil)
		require.EqualError(t, err, status.ErrURIDecoding.Error())
	})

	t.Run("4kb slightly escaped", func(t *testing.T) {
		str := "/" + disperse("%5f", "a", 10, 4095)
		buff := make([]byte, 0, 4096)
		decoded, err := Decode([]byte(str), buff)
		require.NoError(t, err)
		want := "/" + strings.Repeat("_"+strings.Repeat("a", 10), 4095/len("%5f"+strings.Repeat("a", 10)))
		require.Equal(t, want, string(decoded))
		require.Equal(t, 4096, cap(decoded))
	})
}

func BenchmarkDecode(b *testing.B) {
	bench := func(b *testing.B, segment string) {
		str := []byte("/" + strings.Repeat(segment, 4095/len(segment)))
		buff := make([]byte, 0, len(str))
		b.SetBytes(int64(len(str)))
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			_, _ = Decode(str, buff[:0])
		}
	}

	b.Run("4kb unescaped", func(b *testing.B) {
		bench(b, "a")
	})

	b.Run("4kb slightly escaped", func(b *testing.B) {
		// one urlencoded part per 10 decoded characters
		bench(b, "%5faaaaaaaaa")
	})

	b.Run("4kb half escaped", func(b *testing.B) {
		bench(b, "%5fa")
	})

	b.Run("4kb only escaped", func(b *testing.B) {
		bench(b, "%5f")
	})
}

// disperse makes a string, which consists of 1:proportion substrings a and b respectfully.
// Repeating them doesn't always result in exactly desiredLen bytes
func disperse(a, b string, proportion, desiredLen int) string {
	return strings.Repeat(a+strings.Repeat(b, proportion), desiredLen/(len(a)+len(b)*proportion))
}
