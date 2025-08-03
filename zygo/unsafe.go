package zygo

import (
	"reflect"
	"unsafe"
)

// clients must never modify b, runtime/race/crash can result.
func UnsafeStringToByteSlice(s string) []byte {
	p := unsafe.Pointer((*reflect.StringHeader)(unsafe.Pointer(&s)).Data)

	var b []byte
	hdr := (*reflect.SliceHeader)(unsafe.Pointer(&b))
	hdr.Data = uintptr(p)
	hdr.Cap = len(s)
	hdr.Len = len(s)

	return b
}
