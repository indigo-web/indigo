package method

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func BenchmarkMethod(b *testing.B) {
	var parsed Method

	for i := Unknown; i <= Count; i++ {
		b.Run(i.String(), func(b *testing.B) {
			m := i.String()
			b.SetBytes(int64(len(m)))
			b.ResetTimer()

			for j := 0; j < b.N; j++ {
				parsed = Parse(m)
			}
		})
	}

	keepalive(parsed)
}

func keepalive(Method) {}

func TestMethod(t *testing.T) {
	for _, method := range List {
		assert.Equal(t, method.String(), Parse(method.String()).String())
	}
}
