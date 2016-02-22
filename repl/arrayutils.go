package zygo

import "fmt"

func MapArray(env *Glisp, fun *SexpFunction, arr *SexpArray) (Sexp, error) {
	result := make([]Sexp, len(arr.Val))
	var err error

	var firstTyp *RegisteredType
	for i := range arr.Val {
		result[i], err = env.Apply(fun, arr.Val[i:i+1])
		if err != nil {
			return &SexpArray{Val: result, Typ: firstTyp}, err
		}
		if firstTyp == nil {
			firstTyp = result[i].Type()
		}
	}

	return &SexpArray{Val: result, Typ: firstTyp}, nil
}

func ConcatArray(arr *SexpArray, rest []Sexp) (Sexp, error) {
	if arr == nil {
		return SexpNull, fmt.Errorf("ConcatArray called with nil arr")
	}
	var res SexpArray
	res.Val = arr.Val
	for i, x := range rest {
		switch t := x.(type) {
		case *SexpArray:
			res.Val = append(res.Val, t.Val...)
		default:
			return &res, fmt.Errorf("ConcatArray error: %d-th argument "+
				"(0-based) is not an array", i)
		}
	}
	return &res, nil
}
