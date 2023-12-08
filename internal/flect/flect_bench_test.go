package flect

import (
	"testing"
	"unsafe"
)

func BenchmarkModelFiller(b *testing.B) {
	b.Run("full", func(b *testing.B) {
		model := NewModel[myStruct]()
		fields := []Field{
			{"A", uptr(5)},
			{"B", uptr(32769)},
			{"C", uptr(67108864)},
		}
		b.ReportAllocs()
		b.SetBytes(int64(unsafe.Sizeof(myStruct{})))
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			_ = model.Instantiate(fields)
		}
	})

	b.Run("single field", func(b *testing.B) {
		model := NewModel[myStruct]()

		b.ReportAllocs()
		b.SetBytes(int64(unsafe.Sizeof(myStruct{}.C)))
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			_ = model.WriteUInt32(myStruct{}, "C", 67108864)
		}
	})
}
