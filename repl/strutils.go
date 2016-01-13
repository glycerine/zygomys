package zygo

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

func StringUtilFunction(env *Glisp, name string, args []Sexp) (Sexp, error) {
	if len(args) != 1 {
		return SexpNull, WrongNargs
	}
	var s string
	switch str := args[0].(type) {
	case SexpStr:
		s = string(str)
	default:
		return SexpNull, fmt.Errorf("string required, got %T", s)
	}

	switch name {
	case "chomp":
		n := len(s)
		if n > 0 && s[n-1] == '\n' {
			return SexpStr(s[:n-1]), nil
		}
		return SexpStr(s), nil
	}
	return SexpNull, fmt.Errorf("unrecognized command '%s'", name)
}
