package zygo

import "fmt"

func MapArray(env *Glisp, fun SexpFunction, arr SexpArray) (SexpArray, error) {
	result := make([]Sexp, len(arr))
	var err error

	for i := range arr {
		result[i], err = env.Apply(fun, arr[i:i+1])
		if err != nil {
			return SexpArray(result), err
		}
	}

	return SexpArray(result), nil
}

func ConcatArray(arr SexpArray, rest []Sexp) (SexpArray, error) {
	for i, x := range rest {
		switch t := x.(type) {
		case SexpArray:
			arr = append(arr, t...)
		default:
			return arr, fmt.Errorf("ConcatArray error: %d-th argument "+
				"(0-based) is not an array", i)
		}
	}
	return arr, nil
}
