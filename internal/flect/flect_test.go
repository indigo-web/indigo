package main

import (
	"testing"
)

func BenchmarkModelFiller(b *testing.B) {
	model := NewModel[MyStruct]()
	fields := []FieldPair{
		{"A", uptr(1)},
		{"B", uptr(1234567)},
		{"C", uptr(5)},
	}
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = model.New(fields)
	}
}
