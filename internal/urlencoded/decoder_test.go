package urlencoded

import (
	"fmt"
	"github.com/indigo-web/indigo/http/status"
	"github.com/indigo-web/utils/uf"
	"github.com/stretchr/testify/require"
	"math/rand/v2"
	"strings"
	"testing"
)

func testDecoder(t *testing.T, decoder func([]byte, []byte) ([]byte, []byte, error)) {
	t.Run("no escaping", func(t *testing.T) {
		str := []byte("/hello")
		decoded, _, err := decoder(str, []byte{})
		require.NoError(t, err)
		require.Equal(t, "/hello", string(decoded))
	})

	t.Run("corners", func(t *testing.T) {
		str := []byte("%2fhello%2f")
		decoded, _, err := decoder(str, []byte{})
		require.NoError(t, err)
		require.Equal(t, "/hello/", string(decoded))
	})

	t.Run("multiple consecutive", func(t *testing.T) {
		str := []byte("%2f%20hello")
		decoded, _, err := decoder(str, []byte{})
		require.NoError(t, err)
		require.Equal(t, "/ hello", string(decoded))
	})

	t.Run("incomplete sequence", func(t *testing.T) {
		str := []byte("%2")
		_, _, err := decoder(str, []byte{})
		require.EqualError(t, err, status.ErrURLDecoding.Error())
	})

	t.Run("invalid code", func(t *testing.T) {
		str := []byte("%2j")
		_, _, err := decoder(str, []byte{})
		require.EqualError(t, err, status.ErrURLDecoding.Error())
	})

	t.Run("4kb slightly escaped", func(t *testing.T) {
		str := []byte("/" + disperse("%5f", "a", 10, 4095))
		decoded, _, err := decoder(str, []byte{})
		require.NoError(t, err)
		want := "/" + strings.Repeat("_"+strings.Repeat("a", 10), 4095/len("%5f"+strings.Repeat("a", 10)))
		require.Equal(t, want, string(decoded))
		require.Equal(t, 4096, cap(decoded))
	})

	t.Run("decode into itself", func(t *testing.T) {
		for _, tc := range []struct {
			Encoded []byte
			Want    string
		}{
			{[]byte("%2a"), "*"},
			{[]byte("he%6c%6Co"), "hello"},
			{[]byte("nothing here"), "nothing here"},
		} {
			decoded, _, err := decoder(tc.Encoded, tc.Encoded[:0])
			require.NoError(t, err)
			require.Equal(t, tc.Want, string(decoded))
		}
	})
}

func TestDecode(t *testing.T) {
	testDecoder(t, Decode)
}

func TestExtendedDecode(t *testing.T) {
	testDecoder(t, ExtendedDecode)
}

func bench(b *testing.B, name string, decoder func(src, dst []byte) (decoded, buff []byte, err error)) {
	sizes := []int{
		4096,  /*4kb*/
		32768, /*32kb*/
		65536, /*64kb*/
	}
	buff := make([]byte, 0, sizes[len(sizes)-1])
	proportions := []struct{ A, B int }{
		{0, 1},
		{1, 5},
		{3, 5},
		{1, 1},
		{1, 0},
	}

	for _, size := range sizes {
		for _, prop := range proportions {
			b.Run(
				fmt.Sprintf("%s %d bytes %d:%d proportion", name, size, prop.A, prop.B),
				func(b *testing.B) {
					str := uf.S2B(mix("%2a", "a", prop.A, prop.B, size))
					b.ReportAllocs()
					b.SetBytes(int64(len(str)))
					b.ResetTimer()

					for i := 0; i < b.N; i++ {
						_, _, _ = decoder(str, buff)
					}
				},
			)
		}
	}
}

func BenchmarkDecode(b *testing.B) {
	bench(b, "Decode", Decode)
}

func BenchmarkExtendedDecode(b *testing.B) {
	bench(b, "ExtendedDecode", ExtendedDecode)
}

// mix produces a mix of a and b substrings, randomly distributed in respect to given
// proportions over the (not necessarily exactly) `length` bytes
func mix(a, b string, propA, propB, length int) string {
	ratio := length / (len(a)*propA + len(b)*propB)
	as, bs := propA*ratio, propB*ratio
	arr := make([]string, 0, as+bs)

	for range as {
		arr = append(arr, a)
	}
	for range bs {
		arr = append(arr, b)
	}

	rand.Shuffle(len(arr), func(i, j int) {
		arr[i], arr[j] = arr[j], arr[i]
	})

	return strings.Join(arr, "")
}

// disperse returns a string of length `length` or more. Guaranteed to not break the sequences.
// The resulting string consists of substrings `a` and `b`, where the proportion of `a:b` is equal
// to `1:proportion`. Produces a predictable string, which is desired in tests.
func disperse(a, b string, proportion, length int) string {
	return strings.Repeat(
		a+strings.Repeat(b, proportion),
		length/(len(a)+len(b)*proportion),
	)
}
