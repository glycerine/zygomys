package glisp

import (
	"errors"
)

func IsList(expr Sexp) bool {
	if expr == SexpNull {
		return true
	}
	switch list := expr.(type) {
	case SexpPair:
		return IsList(list.tail)
	}
	return false
}

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
