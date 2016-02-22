package zygo

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
	var ia *SexpInt
	var ib *SexpInt

	switch i := a.(type) {
	case *SexpInt:
		ia = i
	case SexpChar:
		ia = &SexpInt{Val: int64(i.Val)}
	default:
		return SexpNull, WrongType
	}

	switch i := b.(type) {
	case *SexpInt:
		ib = i
	case SexpChar:
		ib = &SexpInt{Val: int64(i.Val)}
	default:
		return SexpNull, WrongType
	}

	switch op {
	case ShiftLeft:
		return &SexpInt{Val: ia.Val << uint(ib.Val)}, nil
	case ShiftRightArith:
		return &SexpInt{Val: ia.Val >> uint(ib.Val)}, nil
	case ShiftRightLog:
		return &SexpInt{Val: int64(uint(ia.Val) >> uint(ib.Val))}, nil
	case Modulo:
		return &SexpInt{Val: ia.Val % ib.Val}, nil
	case BitAnd:
		return &SexpInt{Val: ia.Val & ib.Val}, nil
	case BitOr:
		return &SexpInt{Val: ia.Val | ib.Val}, nil
	case BitXor:
		return &SexpInt{Val: ia.Val ^ ib.Val}, nil
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
		return SexpFloat{Val: a.Val + b.Val}
	case Sub:
		return SexpFloat{Val: a.Val - b.Val}
	case Mult:
		return SexpFloat{Val: a.Val * b.Val}
	case Div:
		return SexpFloat{Val: a.Val / b.Val}
	case Pow:
		return SexpFloat{Val: math.Pow(float64(a.Val), float64(b.Val))}
	}
	return SexpNull
}

func NumericIntDo(op NumericOp, a, b *SexpInt) Sexp {
	switch op {
	case Add:
		return &SexpInt{Val: a.Val + b.Val}
	case Sub:
		return &SexpInt{Val: a.Val - b.Val}
	case Mult:
		return &SexpInt{Val: a.Val * b.Val}
	case Div:
		if a.Val%b.Val == 0 {
			return &SexpInt{Val: a.Val / b.Val}
		} else {
			return SexpFloat{Val: float64(a.Val) / float64(b.Val)}
		}
	case Pow:
		return &SexpInt{Val: int64(math.Pow(float64(a.Val), float64(b.Val)))}
	}
	return SexpNull
}

func NumericMatchFloat(op NumericOp, a SexpFloat, b Sexp) (Sexp, error) {
	var fb SexpFloat
	switch tb := b.(type) {
	case SexpFloat:
		fb = tb
	case *SexpInt:
		fb = SexpFloat{Val: float64(tb.Val)}
	case SexpChar:
		fb = SexpFloat{Val: float64(tb.Val)}
	default:
		return SexpNull, WrongType
	}
	return NumericFloatDo(op, a, fb), nil
}

func NumericMatchInt(op NumericOp, a *SexpInt, b Sexp) (Sexp, error) {
	switch tb := b.(type) {
	case SexpFloat:
		return NumericFloatDo(op, SexpFloat{Val: float64(a.Val)}, tb), nil
	case *SexpInt:
		return NumericIntDo(op, a, tb), nil
	case SexpChar:
		return NumericIntDo(op, a, &SexpInt{Val: int64(tb.Val)}), nil
	}
	return SexpNull, WrongType
}

func NumericMatchChar(op NumericOp, a SexpChar, b Sexp) (Sexp, error) {
	var res Sexp
	switch tb := b.(type) {
	case SexpFloat:
		res = NumericFloatDo(op, SexpFloat{Val: float64(a.Val)}, tb)
	case *SexpInt:
		res = NumericIntDo(op, &SexpInt{Val: int64(a.Val)}, tb)
	case SexpChar:
		res = NumericIntDo(op, &SexpInt{Val: int64(a.Val)}, &SexpInt{Val: int64(tb.Val)})
	default:
		return SexpNull, WrongType
	}
	switch tres := res.(type) {
	case SexpFloat:
		return tres, nil
	case *SexpInt:
		return SexpChar{Val: rune(tres.Val)}, nil
	}
	return SexpNull, errors.New("unexpected result")
}

func NumericDo(op NumericOp, a, b Sexp) (Sexp, error) {
	switch ta := a.(type) {
	case SexpFloat:
		return NumericMatchFloat(op, ta, b)
	case *SexpInt:
		return NumericMatchInt(op, ta, b)
	case SexpChar:
		return NumericMatchChar(op, ta, b)
	}
	return SexpNull, WrongType
}
