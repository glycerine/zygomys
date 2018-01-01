package zygo

import (
	"reflect"
)

// true if target is type *T where T
// is a struct/string/int/other-non-pointer type.
func IsExactlySinglePointer(target interface{}) bool {
	//	va, isVa := target.(reflect.Value)
	//	if isVa {
	//		return IsExactlySinglePointerType(va.Type())
	//	}
	typ := reflect.ValueOf(target).Type()
	return IsExactlySinglePointerType(typ)
}
func IsExactlySinglePointerType(typ reflect.Type) bool {
	kind := typ.Kind()
	if kind != reflect.Ptr {
		return false
	}
	typ2 := typ.Elem()
	kind2 := typ2.Kind()
	if kind2 == reflect.Ptr {
		return false // two level pointer
	}
	return true
}

// true if target is of type **T where T is
// a struct/string/int/other-non-pointer type.
func IsExactlyDoublePointer(target interface{}) bool {
	typ := reflect.ValueOf(target).Type()
	kind := typ.Kind()
	if kind != reflect.Ptr {
		return false
	}
	typ2 := typ.Elem()
	kind2 := typ2.Kind()
	if kind2 != reflect.Ptr {
		return false
	}
	if typ2.Elem().Kind() == reflect.Ptr {
		return false // triple level pointer, not double.
	}
	return true
}

func PointerDepth(typ reflect.Type) int {
	return pointerDepthHelper(typ, 0)
}

func pointerDepthHelper(typ reflect.Type, accum int) int {
	kind := typ.Kind()
	if kind != reflect.Ptr {
		return accum
	}
	return pointerDepthHelper(typ.Elem(), accum+1)
}
