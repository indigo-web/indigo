package method

import "testing"

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
