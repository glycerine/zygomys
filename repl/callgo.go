package zygo

import (
	"fmt"
	"github.com/shurcooL/go-goon"
)

// demostrate calling Go

// Using reflection, invoke a Go method on a struct or interface.
// args[0] is a hash with an an attached GoStruct
// args[1] is a hash representing a method call on that struct.
// The returned Sexp is a hash that represents the result of that call.
func CallGo(env *Glisp, name string, args []Sexp) (Sexp, error) {
	if len(args) < 2 {
		return SexpNull, WrongNargs
	}
	object, isHash := args[0].(SexpHash)
	if !isHash {
		return SexpNull, fmt.Errorf("CallGo() error: first argument must be a hash or defmap with an attached GoObject")
	}
	method, isHash := args[1].(SexpHash)
	if !isHash {
		return SexpNull, fmt.Errorf("TwoHashCallGo() error: second argument must be a hash/record representing the method to call")
	}

	goon.Dump(object)
	goon.Dump(method)
	//fmt.Printf("expected = '%#v'\n", expected)

	// construct the call by reflection

	// return the result in a hash record too.
	return SexpNull, nil
}
