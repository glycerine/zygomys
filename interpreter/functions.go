package glisp

import (
	"errors"
	"fmt"
	"bytes"
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

func signum(f SexpFloat) int {
	if f > 0 {
		return 1
	}
	if f < 0 {
		return -1
	}
	return 0
}

func compareFloat(f SexpFloat, expr Sexp) (int, error) {
	switch e := expr.(type) {
	case SexpInt:
		return signum(f - SexpFloat(e)), nil
	case SexpFloat:
		return signum(f - e), nil
	case SexpChar:
		return signum(f - SexpFloat(e)), nil
	}
	errmsg := fmt.Sprintf("cannot compare %T to %T", f, expr)
	return 0, errors.New(errmsg)
}

func compareInt(i SexpInt, expr Sexp) (int, error) {
	switch e := expr.(type) {
	case SexpInt:
		return signum(SexpFloat(i - e)), nil
	case SexpFloat:
		return signum(SexpFloat(i) - e), nil
	case SexpChar:
		return signum(SexpFloat(i - SexpInt(e))), nil
	}
	errmsg := fmt.Sprintf("cannot compare %T to %T", i, expr)
	return 0, errors.New(errmsg)
}

func compareChar(c SexpChar, expr Sexp) (int, error) {
	switch e := expr.(type) {
	case SexpInt:
		return signum(SexpFloat(SexpInt(c) - e)), nil
	case SexpFloat:
		return signum(SexpFloat(c) - e), nil
	case SexpChar:
		return signum(SexpFloat(c - e)), nil
	}
	errmsg := fmt.Sprintf("cannot compare %T to %T", c, expr)
	return 0, errors.New(errmsg)
}

func compareString(s SexpStr, expr Sexp) (int, error) {
	switch e := expr.(type) {
	case SexpStr:
		return bytes.Compare([]byte(s), []byte(e)), nil
	}
	errmsg := fmt.Sprintf("cannot compare %T to %T", s, expr)
	return 0, errors.New(errmsg)
}

func Compare(a Sexp, b Sexp) (int, error) {
	switch at := a.(type) {
	case SexpInt:
		return compareInt(at, b)
	case SexpChar:
		return compareChar(at, b)
	case SexpFloat:
		return compareFloat(at, b)
	case SexpStr:
		return compareString(at, b)
	}
	errmsg := fmt.Sprintf("cannot compare %T to %T", a, b)
	return 0, errors.New(errmsg)
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
