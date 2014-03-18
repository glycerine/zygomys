package glisp

import (
	"errors"
)

func ConcatStr(str SexpStr, expr Sexp) (SexpStr, error) {
	var str2 SexpStr
	switch t := expr.(type) {
	case SexpStr:
		str2 = t
	default:
		return SexpStr(""), errors.New("second argument is not a string")
	}

	return str + str2, nil
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
