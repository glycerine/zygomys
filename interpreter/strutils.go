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
