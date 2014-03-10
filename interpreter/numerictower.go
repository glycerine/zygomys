package glisp

import (
	"errors"
)

type BinaryIntOp int
const (
	ShiftLeft BinaryIntOp = iota
	ShiftRightArith
	ShiftRightLog
	Modulo
)

var WrongType error = errors.New("operands have invalid type")

func BinaryIntDo(op BinaryIntOp, a, b Sexp) (Sexp, error) {
	var ia SexpInt
	var ib SexpInt

	switch i := a.(type) {
	case SexpInt:
		ia = i
	default:
		return SexpNull, WrongType
	}

	switch i := b.(type) {
	case SexpInt:
		ib = i
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
	}
	return SexpNull, errors.New("unrecognized shift operation")
}

type NumericOp int
const (
	Add NumericOp = iota
	Sub
	Mult
	Div
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
		if a % b == 0 {
			return a / b
		} else {
			return SexpFloat(a) / SexpFloat(b)
		}
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
