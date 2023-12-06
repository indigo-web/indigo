package main

import (
	"fmt"
	"reflect"
	"unsafe"
)

type MyStruct struct {
	A uint8
	B uint64
	C uint8
}

func (m MyStruct) String() string {
	return fmt.Sprintf("MyStruct{%d, %d, %d}", m.A, m.B, m.C)
}

type fieldData struct {
	Size, Offset uintptr
}

type Model[T any] struct {
	offsets     map[string]fieldData
	offsetsInts []fieldData
}

func NewModel[T any]() Model[T] {
	var zero [0]T
	typ := reflect.TypeOf(zero).Elem()
	if typ.Kind() != reflect.Struct {
		panic("not a struct")
	}

	model := Model[T]{
		offsets: map[string]fieldData{},
	}

	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		model.offsetsInts = append(model.offsetsInts, fieldData{
			Size:   field.Type.Size(),
			Offset: field.Offset,
		})
		model.offsets[field.Name] = fieldData{
			Size:   field.Type.Size(),
			Offset: field.Offset,
		}
	}

	return model
}

type FieldPair struct {
	Key  string
	Data unsafe.Pointer
}

func (m Model[T]) WriteUInt8(into T, key string, num uint8) T {
	field := m.offsets[key]
	memcpy(unsafe.Add(unsafe.Pointer(&into), field.Offset), unsafe.Pointer(&num), field.Size)
	return into
}

func (m Model[T]) New(values []FieldPair) T {
	var zero T
	zeroUPtr := unsafe.Pointer(&zero)

	for _, value := range values {
		field := m.offsets[value.Key]
		fieldPtr := unsafe.Add(zeroUPtr, field.Offset)
		memcpy(fieldPtr, value.Data, field.Size)
	}

	return zero
}

func memcpy(dst, src unsafe.Pointer, size uintptr) {
	copy(unsafe.Slice((*byte)(dst), size), unsafe.Slice((*byte)(src), size))
}

func main() {
	model := NewModel[MyStruct]()
	fmt.Println("model:", model)
	m := MyStruct{}
	updated := model.WriteUInt8(m, "A", 87)
	fmt.Println(updated.String())
	filled := model.New([]FieldPair{
		{"A", uptr(1)},
		{"B", uptr(1234567)},
		{"C", uptr(5)},
	})
	fmt.Println(filled.String())
}

func uptr[T any](val T) unsafe.Pointer {
	return unsafe.Pointer(&val)
}
