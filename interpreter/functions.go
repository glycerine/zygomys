package glisp

import (
	"errors"
	"fmt"
	"bytes"
)

var WrongNargs error = errors.New("wrong number of arguments")

type GlispFunction func(*Glisp, SexpSymbol, int) error
type GlispUserFunction func(*Glisp, string, []Sexp) (Sexp, error)

func MakeUserFunction(fun GlispUserFunction) GlispFunction {
	return func(glisp *Glisp, sym SexpSymbol, nargs int) error {
		arr, err := glisp.datastack.PopExpressions(nargs)
		if err != nil {
			return err
		}
		res, err := fun(glisp, sym.name, arr)
		if err != nil {
			return err
		}
		glisp.datastack.PushExpr(res)
		return nil
	}
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
	case SexpUint:
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
	case SexpUint:
		return signum(SexpFloat(i - SexpInt(e))), nil
	case SexpFloat:
		return signum(SexpFloat(i) - e), nil
	case SexpChar:
		return signum(SexpFloat(i - SexpInt(e))), nil
	}
	errmsg := fmt.Sprintf("cannot compare %T to %T", i, expr)
	return 0, errors.New(errmsg)
}

func compareUint(u SexpUint, expr Sexp) (int, error) {
	switch e := expr.(type) {
	case SexpInt:
		return signum(SexpFloat(SexpInt(u) - e)), nil
	case SexpUint:
		return signum(SexpFloat(u - e)), nil
	case SexpFloat:
		return signum(SexpFloat(u) - e), nil
	case SexpChar:
		return signum(SexpFloat(u - SexpUint(e))), nil
	}
	errmsg := fmt.Sprintf("cannot compare %T to %T", u, expr)
	return 0, errors.New(errmsg)
}

func compareChar(c SexpChar, expr Sexp) (int, error) {
	switch e := expr.(type) {
	case SexpInt:
		return signum(SexpFloat(SexpInt(c) - e)), nil
	case SexpUint:
		return signum(SexpFloat(SexpUint(c) - e)), nil
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
	case SexpUint:
		return compareUint(at, b)
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

func CompareFunction(glisp *Glisp, sym SexpSymbol, nargs int) error {
	if nargs != 2 {
		return WrongNargs
	}
	arr, err := glisp.datastack.PopExpressions(nargs)
	if err != nil {
		return err
	}
	res, err := Compare(arr[0], arr[1])
	if err != nil {
		return err
	}

	cond := false
	switch sym.name {
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
	glisp.datastack.PushExpr(SexpBool(cond))

	return nil
}

/*func ArithFunction(glisp *Glisp, sym SexpSymbol, nargs int) error {
	arr, err := glisp.datastack.PopExpressions(nargs)
	if err != nil {
		return err
	}
}*/

var BuiltinFunctions = map[string]GlispFunction {
	"<" : CompareFunction,
	">" : CompareFunction,
	"<=": CompareFunction,
	">=": CompareFunction,
	"=" : CompareFunction,
	"not=": CompareFunction,
}
