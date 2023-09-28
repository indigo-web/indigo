package coding

import (
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
)

const megabyte = 1024 * 1024

func BenchmarkGZIP_Encode1mb(b *testing.B) {
	repeatedSequence := "qgkvasdf sdfjghsdjfas sjkdfhjksd"
	data := []byte(strings.Repeat(repeatedSequence, megabyte/len(repeatedSequence)))
	coder := NewGZIP(make([]byte, megabyte))
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = coder.Encode(data)
	}
}

func BenchmarkGZIP_Decode1mb(b *testing.B) {
	repeatedSequence := "qgkvasdf sdfjghsdjfas sjkdfhjksd"
	data := []byte(strings.Repeat(repeatedSequence, megabyte/len(repeatedSequence)))
	coder := NewGZIP(make([]byte, megabyte))
	compressedData, err := coder.Encode(data)
	require.NoError(b, err)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = coder.Decode(compressedData)
	}
}
