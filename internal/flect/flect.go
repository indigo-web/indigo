package flect

import (
	"reflect"
	"unsafe"
)

type Field struct {
	Key  string
	Data unsafe.Pointer
}

type fieldData struct {
	Size, Offset uintptr
}

type Model[T any] struct {
	offsets *attrsMap
}

func NewModel[T any]() Model[T] {
	var zero [0]T
	typ := reflect.TypeOf(zero).Elem()
	if typ.Kind() != reflect.Struct {
		panic("not a struct")
	}

	model := Model[T]{
		offsets: new(attrsMap),
	}

	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		model.offsets.Insert(field.Name, fieldData{
			Size:   field.Type.Size(),
			Offset: field.Offset,
		})
	}

	return model
}

func (m Model[T]) Instantiate(values []Field) T {
	var zero T
	zeroUPtr := unsafe.Pointer(&zero)

	for _, value := range values {
		field, found := m.offsets.Lookup(value.Key)
		if !found {
			continue
		}
		fieldPtr := unsafe.Add(zeroUPtr, field.Offset)
		memcpy(fieldPtr, value.Data, field.Size)
	}

	return zero
}

func (m Model[T]) write(into T, key string, val unsafe.Pointer, size uintptr) T {
	field, found := m.offsets.Lookup(key)
	if !found {
		return into
	}
	memcpy(unsafe.Add(unsafe.Pointer(&into), field.Offset), val, size)
	return into
}

func (m Model[T]) WriteUInt8(into T, key string, num uint8) T {
	return m.write(into, key, unsafe.Pointer(&num), unsafe.Sizeof(num))
}

func (m Model[T]) WriteUInt16(into T, key string, num uint16) T {
	return m.write(into, key, unsafe.Pointer(&num), unsafe.Sizeof(num))
}

func (m Model[T]) WriteUInt32(into T, key string, num uint32) T {
	return m.write(into, key, unsafe.Pointer(&num), unsafe.Sizeof(num))
}

func (m Model[T]) WriteUInt64(into T, key string, num uint64) T {
	return m.write(into, key, unsafe.Pointer(&num), unsafe.Sizeof(num))
}

func (m Model[T]) WriteInt8(into T, key string, num int8) T {
	return m.write(into, key, unsafe.Pointer(&num), unsafe.Sizeof(num))
}

func (m Model[T]) WriteInt16(into T, key string, num int16) T {
	return m.write(into, key, unsafe.Pointer(&num), unsafe.Sizeof(num))
}

func (m Model[T]) WriteInt32(into T, key string, num int32) T {
	return m.write(into, key, unsafe.Pointer(&num), unsafe.Sizeof(num))
}

func (m Model[T]) WriteInt64(into T, key string, num int64) T {
	return m.write(into, key, unsafe.Pointer(&num), unsafe.Sizeof(num))
}

func (m Model[T]) WriteString(into T, key string, value string) T {
	return m.write(into, key, unsafe.Pointer(&value), unsafe.Sizeof(value))
}

func memcpy(dst, src unsafe.Pointer, size uintptr) {
	copy(unsafe.Slice((*byte)(dst), size), unsafe.Slice((*byte)(src), size))
}
