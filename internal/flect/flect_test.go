package flect

import (
	"github.com/stretchr/testify/require"
	"testing"
	"unsafe"
)

type myStruct struct {
	A byte
	B uint16
	C int32
}

func TestFlect(t *testing.T) {
	model := NewModel[myStruct]()

	t.Run("instantiate", func(t *testing.T) {
		m := model.Instantiate([]Field{
			{"A", uptr(5)},
			{"B", uptr(32769)},
			{"C", uptr(-67108864)},
		})

		require.Equal(t, uint8(5), m.A)
		require.Equal(t, uint16(32769), m.B)
		require.Equal(t, int32(-67108864), m.C)
	})

	t.Run("fill partially", func(t *testing.T) {
		m := model.Instantiate([]Field{
			{"A", uptr(5)},
			{"C", uptr(-67108864)},
		})

		require.Equal(t, uint8(5), m.A)
		require.Equal(t, uint16(0), m.B)
		require.Equal(t, int32(-67108864), m.C)
	})

	t.Run("fill with unknown field", func(t *testing.T) {
		m := model.Instantiate([]Field{
			{"A", uptr(5)},
			{"G", uptr(123)},
			{"C", uptr(-67108864)},
			{"M", uptr(123)},
		})

		require.Equal(t, uint8(5), m.A)
		require.Equal(t, uint16(0), m.B)
		require.Equal(t, int32(-67108864), m.C)
	})
}

func uptr[T any](val T) unsafe.Pointer {
	// I'm not actually sure, whether taking a pointer directly here won't result
	// in values, which may be overridden on a next call
	return unsafe.Pointer(ptr(val))
}

func ptr[T any](val T) *T {
	return &val
}
