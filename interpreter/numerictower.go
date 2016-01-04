package glisp

import (
	"errors"
	"math"
)

type IntegerOp int

const (
	ShiftLeft IntegerOp = iota
	ShiftRightArith
	ShiftRightLog
	Modulo
	BitAnd
	BitOr
	BitXor
)

var WrongType error = errors.New("operands have invalid type")

func IntegerDo(op IntegerOp, a, b Sexp) (Sexp, error) {
	var ia SexpInt
	var ib SexpInt

	switch i := a.(type) {
	case SexpInt:
		ia = i
	case SexpChar:
		ia = SexpInt(i)
	default:
		return SexpNull, WrongType
	}

	switch i := b.(type) {
	case SexpInt:
		ib = i
	case SexpChar:
		ib = SexpInt(i)
	default:
		return SexpNull, WrongType
	}

	switch op {
	case ShiftLeft:
		return ia << uint(ib), nil
	case ShiftRightArith:
		return ia >> uint(ib), nil
	case ShiftRightLog:
		return SexpInt(uint(ia) >> uint(ib)), nil
	case Modulo:
		return ia % ib, nil
	case BitAnd:
		return ia & ib, nil
	case BitOr:
		return ia | ib, nil
	case BitXor:
		return ia ^ ib, nil
	}
	return SexpNull, errors.New("unrecognized shift operation")
}

type NumericOp int

const (
	Add NumericOp = iota
	Sub
	Mult
	Div
	Pow
)

func NumericFloatDo(op NumericOp, a, b SexpFloat) Sexp {
	switch op {
	case Add:
		return a + b
	case Sub:
		return a - b
	case Mult:
		return a * b
	case Div:
		return a / b
	case Pow:
		return SexpFloat(math.Pow(float64(a), float64(b)))
	}
	return SexpNull
}

func NumericIntDo(op NumericOp, a, b SexpInt) Sexp {
	switch op {
	case Add:
		return a + b
	case Sub:
		return a - b
	case Mult:
		return a * b
	case Div:
		if a%b == 0 {
			return a / b
		} else {
			return SexpFloat(a) / SexpFloat(b)
		}
	case Pow:
		return SexpInt(math.Pow(float64(a), float64(b)))
	}
	return SexpNull
}

func NumericMatchFloat(op NumericOp, a SexpFloat, b Sexp) (Sexp, error) {
	var fb SexpFloat
	switch tb := b.(type) {
	case SexpFloat:
		fb = tb
	case SexpInt:
		fb = SexpFloat(tb)
	case SexpChar:
		fb = SexpFloat(tb)
	default:
		return SexpNull, WrongType
	}
	return NumericFloatDo(op, a, fb), nil
}

func NumericMatchInt(op NumericOp, a SexpInt, b Sexp) (Sexp, error) {
	switch tb := b.(type) {
	case SexpFloat:
		return NumericFloatDo(op, SexpFloat(a), tb), nil
	case SexpInt:
		return NumericIntDo(op, a, tb), nil
	case SexpChar:
		return NumericIntDo(op, a, SexpInt(tb)), nil
	}
	return SexpNull, WrongType
}

func NumericMatchChar(op NumericOp, a SexpChar, b Sexp) (Sexp, error) {
	var res Sexp
	switch tb := b.(type) {
	case SexpFloat:
		res = NumericFloatDo(op, SexpFloat(a), tb)
	case SexpInt:
		res = NumericIntDo(op, SexpInt(a), tb)
	case SexpChar:
		res = NumericIntDo(op, SexpInt(a), SexpInt(tb))
	default:
		return SexpNull, WrongType
	}
	switch tres := res.(type) {
	case SexpFloat:
		return tres, nil
	case SexpInt:
		return SexpChar(tres), nil
	}
	return SexpNull, errors.New("unexpected result")
}

func NumericDo(op NumericOp, a, b Sexp) (Sexp, error) {
	switch ta := a.(type) {
	case SexpFloat:
		return NumericMatchFloat(op, ta, b)
	case SexpInt:
		return NumericMatchInt(op, ta, b)
	case SexpChar:
		return NumericMatchChar(op, ta, b)
	}
	return SexpNull, WrongType
}
