package glisp

import (
	"fmt"
	"errors"
	"bytes"
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


