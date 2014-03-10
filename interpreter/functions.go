package glisp

import (
	"errors"
)

var WrongNargs error = errors.New("wrong number of arguments")

type GlispFunction []Instruction
type GlispUserFunction func(*Glisp, string, []Sexp) (Sexp, error)

func (f GlispFunction) SexpString() string {
	return "function"
}

func (f GlispUserFunction) SexpString() string {
	return "user_function"
}

func CompareFunction(glisp *Glisp, name string, args []Sexp) (Sexp, error) {
	if len(args) != 2 {
		return SexpNull, WrongNargs
	}

	res, err := Compare(args[0], args[1])
	if err != nil {
		return SexpNull, err
	}

	cond := false
	switch name {
	case "<":
		cond = res < 0
	case ">":
		cond = res > 0
	case "<=":
		cond = res <= 0
	case ">=":
		cond = res >= 0
	case "=":
		cond = res == 0
	case "not=":
		cond = res != 0
	}

	return SexpBool(cond), nil
}

/*func ArithFunction(glisp *Glisp, sym SexpSymbol, nargs int) error {
	arr, err := glisp.datastack.PopExpressions(nargs)
	if err != nil {
		return err
	}
}*/

var BuiltinFunctions = map[string]GlispUserFunction {
	"<" : CompareFunction,
	">" : CompareFunction,
	"<=": CompareFunction,
	">=": CompareFunction,
	"=" : CompareFunction,
	"not=": CompareFunction,
}
