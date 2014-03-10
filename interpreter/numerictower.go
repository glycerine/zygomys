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

func BinaryIntDo(op BinaryIntOp, a, b Sexp) (Sexp, error) {
	var ia SexpInt
	var ib SexpInt
	wrongtype := errors.New("operands must be integers")

	switch i := a.(type) {
	case SexpInt:
		ia = i
	default:
		return SexpNull, wrongtype
	}

	switch i := b.(type) {
	case SexpInt:
		ib = i
	default:
		return SexpNull, wrongtype
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

type ArithOp int
const (
	ArithAdd ArithOp = iota
	ArithSub
	ArithMult
	ArithDiv
)
