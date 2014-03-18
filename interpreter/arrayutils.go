package glisp

import (
	"errors"
)

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

func ConcatArray(arr SexpArray, expr Sexp) (SexpArray, error) {
	var arr2 SexpArray
	switch t := expr.(type) {
	case SexpArray:
		arr2 = t
	default:
		return arr, errors.New("second argument is not an array")
	}

	return append(arr, arr2...), nil
}
