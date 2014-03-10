package glisp

import (
	"errors"
)

type ShiftOp int
const (
	ShiftLeft ShiftOp = iota
	ShiftRightArith
	ShiftRightLog
)

func Shift(op ShiftOp, a, b Sexp) (Sexp, error) {
	var ia SexpInt
	var ib SexpInt
	wrongtype := errors.New("shift operands must be integers")

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
	}
	return SexpNull, errors.New("unrecognized shift operation")
}
