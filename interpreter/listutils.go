package glisp

import (
	"errors"
)

var NotAList = errors.New("not a list")

func ListToArray(expr Sexp) ([]Sexp, error) {
	if !IsList(expr) {
		return nil, NotAList
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

func MapList(env *Glisp, fun SexpFunction, expr Sexp) (Sexp, error) {
	if expr == SexpNull {
		return SexpNull, nil
	}

	var list SexpPair
	switch e := expr.(type) {
	case SexpPair:
		list = e
	default:
		return SexpNull, NotAList
	}

	var err error

	list.head, err = env.Apply(fun, []Sexp{list.head})

	if err != nil {
		return SexpNull, err
	}

	list.tail, err = MapList(env, fun, list.tail)

	if err != nil {
		return SexpNull, err
	}

	return list, nil
}

func ConcatList(a SexpPair, b Sexp) (Sexp, error) {
	if !IsList(b) {
		return SexpNull, NotAList
	}

	if a.tail == SexpNull {
		return SexpPair{a.head, b}, nil
	}

	switch t := a.tail.(type) {
	case SexpPair:
		newtail, err := ConcatList(t, b)
		if err != nil {
			return SexpNull, err
		}
		return SexpPair{a.head, newtail}, nil
	}

	return SexpNull, NotAList
}
