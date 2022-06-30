package internal

import (
	"reflect"
	"unsafe"
)

/*
S2B https://github.com/valyala/fasthttp#tricks-with-byte-buffers
*/
func S2B(s string) (b []byte) {
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
B2S same as B2S, but does the opposite, also described in link above
*/
func B2S(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}
