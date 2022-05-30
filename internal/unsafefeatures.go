package internal

import (
	"reflect"
	"unsafe"
)

/*
https://github.com/valyala/fasthttp#tricks-with-byte-buffers
*/
func s2b(s string) (b []byte) {
	/* #nosec G103 */
	bh := (*reflect.SliceHeader)(unsafe.Pointer(&b))
	/* #nosec G103 */
	sh := (*reflect.StringHeader)(unsafe.Pointer(&s))
	bh.Data = sh.Data
	bh.Cap = sh.Len
	bh.Len = sh.Len

	return b
}

/*
same as s2b, but does the opposite, also described in link above
*/
func b2s(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}
