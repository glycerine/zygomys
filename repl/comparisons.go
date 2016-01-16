package zygo

import (
	"bytes"
	"errors"
	"fmt"
)

func signumFloat(f SexpFloat) int {
	if f > 0 {
		return 1
	}
	if f < 0 {
		return -1
	}
	return 0
}

func signumInt(i SexpInt) int {
	if i > 0 {
		return 1
	}
	if i < 0 {
		return -1
	}
	return 0
}

func compareFloat(f SexpFloat, expr Sexp) (int, error) {
	switch e := expr.(type) {
	case SexpInt:
		return signumFloat(f - SexpFloat(e)), nil
	case SexpFloat:
		return signumFloat(f - e), nil
	case SexpChar:
		return signumFloat(f - SexpFloat(e)), nil
	}
	errmsg := fmt.Sprintf("cannot compare %T to %T", f, expr)
	return 0, errors.New(errmsg)
}

func compareInt(i SexpInt, expr Sexp) (int, error) {
	switch e := expr.(type) {
	case SexpInt:
		return signumInt(i - e), nil
	case SexpFloat:
		return signumFloat(SexpFloat(i) - e), nil
	case SexpChar:
		return signumInt(i - SexpInt(e)), nil
	}
	errmsg := fmt.Sprintf("cannot compare %T to %T", i, expr)
	return 0, errors.New(errmsg)
}

func compareChar(c SexpChar, expr Sexp) (int, error) {
	switch e := expr.(type) {
	case SexpInt:
		return signumInt(SexpInt(c) - e), nil
	case SexpFloat:
		return signumFloat(SexpFloat(c) - e), nil
	case SexpChar:
		return signumInt(SexpInt(c - e)), nil
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

func compareSymbol(sym SexpSymbol, expr Sexp) (int, error) {
	switch e := expr.(type) {
	case SexpSymbol:
		return signumInt(SexpInt(sym.number - e.number)), nil
	}
	errmsg := fmt.Sprintf("cannot compare %T to %T", sym, expr)
	return 0, errors.New(errmsg)
}

func comparePair(a SexpPair, b Sexp) (int, error) {
	var bp SexpPair
	switch t := b.(type) {
	case SexpPair:
		bp = t
	default:
		errmsg := fmt.Sprintf("cannot compare %T to %T", a, b)
		return 0, errors.New(errmsg)
	}
	res, err := Compare(a.Head, bp.Head)
	if err != nil {
		return 0, err
	}
	if res != 0 {
		return res, nil
	}
	return Compare(a.Tail, bp.Tail)
}

func compareArray(a SexpArray, b Sexp) (int, error) {
	var ba SexpArray
	switch t := b.(type) {
	case SexpArray:
		ba = t
	default:
		errmsg := fmt.Sprintf("cannot compare %T to %T", a, b)
		return 0, errors.New(errmsg)
	}
	var length int
	if len(a) < len(ba) {
		length = len(a)
	} else {
		length = len(ba)
	}

	for i := 0; i < length; i++ {
		res, err := Compare(a[i], ba[i])
		if err != nil {
			return 0, err
		}
		if res != 0 {
			return res, nil
		}
	}

	return signumInt(SexpInt(len(a) - len(ba))), nil
}

func compareBool(a SexpBool, b Sexp) (int, error) {
	var bb SexpBool
	switch bt := b.(type) {
	case SexpBool:
		bb = bt
	default:
		errmsg := fmt.Sprintf("cannot compare %T to %T", a, b)
		return 0, errors.New(errmsg)
	}

	// true > false
	if a && bb {
		return 0, nil
	}
	if a {
		return 1, nil
	}
	if bb {
		return -1, nil
	}
	return 0, nil
}

func Compare(a Sexp, b Sexp) (int, error) {
	switch at := a.(type) {
	case SexpInt:
		return compareInt(at, b)
	case SexpChar:
		return compareChar(at, b)
	case SexpFloat:
		return compareFloat(at, b)
	case SexpBool:
		return compareBool(at, b)
	case SexpStr:
		return compareString(at, b)
	case SexpSymbol:
		return compareSymbol(at, b)
	case SexpPair:
		return comparePair(at, b)
	case SexpArray:
		return compareArray(at, b)
	case SexpSentinel:
		if at == SexpNull && b == SexpNull {
			return 0, nil
		} else {
			return -1, nil
		}
	}
	errmsg := fmt.Sprintf("cannot compare %T to %T", a, b)
	return 0, errors.New(errmsg)
}
