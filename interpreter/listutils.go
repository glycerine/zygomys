package glisp

import (
	"errors"
)

func ListToArray(expr Sexp) ([]Sexp, error) {
	if !IsList(expr) {
		return nil, errors.New("not a list")
	}
	arr := make([]Sexp, 0)

	for expr != SexpNull {
		list := expr.(SexpPair)
		arr = append(arr, list.head)
		expr = list.tail
	}

	return arr, nil
}

func MakeList(expressions []Sexp) Sexp {
	if len(expressions) == 0 {
		return SexpNull
	}

	return SexpPair{expressions[0], MakeList(expressions[1:])}
}
