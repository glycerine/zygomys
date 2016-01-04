package glisp

import (
	"errors"
	"fmt"
)

func ConcatStr(str SexpStr, rest []Sexp) (SexpStr, error) {
	res := str
	for i, x := range rest {
		switch t := x.(type) {
		case SexpStr:
			res = res + t
		default:
			return SexpStr(""), fmt.Errorf("ConcatStr error: %d-th argument (0-based) is not a string", i)
		}
	}

	return SexpStr(res), nil
}

func AppendStr(str SexpStr, expr Sexp) (SexpStr, error) {
	var chr SexpChar
	switch t := expr.(type) {
	case SexpChar:
		chr = t
	default:
		return SexpStr(""), errors.New("second argument is not a char")
	}

	return str + SexpStr(chr), nil
}
