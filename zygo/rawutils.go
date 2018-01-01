package zygo

import (
	"errors"
	"fmt"
)

func MakeRaw(args []Sexp) (*SexpRaw, error) {
	raw := make([]byte, 0)
	for i := 0; i < len(args); i++ {
		switch e := args[i].(type) {
		case *SexpStr:
			a := []byte(e.S)
			raw = append(raw, a...)
		default:
			return &SexpRaw{},
				fmt.Errorf("raw takes only string arguments. We see %T: '%v'", e, e)
		}
	}
	return &SexpRaw{Val: raw}, nil
}

func RawToStringFunction(env *Zlisp, name string, args []Sexp) (Sexp, error) {
	if len(args) != 1 {
		return SexpNull, WrongNargs
	}

	switch t := args[0].(type) {
	case *SexpRaw:
		return &SexpStr{S: string(t.Val)}, nil
	}
	return SexpNull, errors.New("argument must be raw")
}
