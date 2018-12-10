package zygo

import (
	"errors"
	"fmt"
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
	case *SexpUint64:
		return UintegerDo(op, i, b)
	case *SexpChar:
		ia = &SexpInt{Val: int64(i.Val)}
	default:
		return SexpNull, WrongType
	}

	switch i := b.(type) {
	case *SexpInt:
		ib = i
	case *SexpUint64:
		return UintegerDo(op, &SexpUint64{Val: uint64(ia.Val)}, b)
	case *SexpChar:
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

func UintegerDo(op IntegerOp, ia *SexpUint64, b Sexp) (Sexp, error) {
	var ib *SexpUint64

	switch i := b.(type) {
	case *SexpUint64:
		ib = i
	case *SexpInt:
		ib = &SexpUint64{Val: uint64(i.Val)}
	case *SexpChar:
		ib = &SexpUint64{Val: uint64(i.Val)}
	default:
		return SexpNull, WrongType
	}

	switch op {
	case ShiftLeft:
		return &SexpUint64{Val: ia.Val << ib.Val}, nil
	case ShiftRightArith:
		return &SexpUint64{Val: ia.Val >> ib.Val}, nil
	case ShiftRightLog:
		return &SexpUint64{Val: ia.Val >> ib.Val}, nil
	case Modulo:
		return &SexpUint64{Val: ia.Val % ib.Val}, nil
	case BitAnd:
		return &SexpUint64{Val: ia.Val & ib.Val}, nil
	case BitOr:
		return &SexpUint64{Val: ia.Val | ib.Val}, nil
	case BitXor:
		return &SexpUint64{Val: ia.Val ^ ib.Val}, nil
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

func NumericFloatDo(op NumericOp, a, b *SexpFloat) Sexp {
	fmt.Printf("top of NumericFloatDo\n")
	switch op {
	case Add:
		return &SexpFloat{Val: a.Val + b.Val}
	case Sub:
		return &SexpFloat{Val: a.Val - b.Val}
	case Mult:
		return &SexpFloat{Val: a.Val * b.Val}
	case Div:
		return &SexpFloat{Val: a.Val / b.Val}
	case Pow:
		return &SexpFloat{Val: math.Pow(float64(a.Val), float64(b.Val))}
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
			return &SexpFloat{Val: float64(a.Val) / float64(b.Val)}
		}
	case Pow:
		return &SexpInt{Val: int64(math.Pow(float64(a.Val), float64(b.Val)))}
	}
	return SexpNull
}

func NumericUint64Do(op NumericOp, a, b *SexpUint64) Sexp {
	fmt.Printf("top of NumericUint64Do\n")
	switch op {
	case Add:
		return &SexpUint64{Val: a.Val + b.Val}
	case Sub:
		return &SexpUint64{Val: a.Val - b.Val}
	case Mult:
		return &SexpUint64{Val: a.Val * b.Val}
	case Div:
		if a.Val%b.Val == 0 {
			return &SexpUint64{Val: a.Val / b.Val}
		} else {
			return &SexpFloat{Val: float64(a.Val) / float64(b.Val)}
		}
	case Pow:
		return &SexpUint64{Val: uint64(math.Pow(float64(a.Val), float64(b.Val)))}
	}
	return SexpNull
}

func NumericMatchFloat(op NumericOp, a *SexpFloat, b Sexp) (Sexp, error) {
	var fb *SexpFloat
	switch tb := b.(type) {
	case *SexpFloat:
		fb = tb
	case *SexpInt:
		fb = &SexpFloat{Val: float64(tb.Val)}
	case *SexpUint64:
		fb = &SexpFloat{Val: float64(tb.Val)}
	case *SexpChar:
		fb = &SexpFloat{Val: float64(tb.Val)}
	default:
		return SexpNull, WrongType
	}
	return NumericFloatDo(op, a, fb), nil
}

func NumericMatchInt(op NumericOp, a *SexpInt, b Sexp) (Sexp, error) {
	switch tb := b.(type) {
	case *SexpFloat:
		return NumericFloatDo(op, &SexpFloat{Val: float64(a.Val)}, tb), nil
	case *SexpInt:
		return NumericIntDo(op, a, tb), nil
	case *SexpUint64:
		return NumericUint64Do(op, &SexpUint64{Val: uint64(a.Val)}, tb), nil
	case *SexpChar:
		return NumericIntDo(op, a, &SexpInt{Val: int64(tb.Val)}), nil
	}
	return SexpNull, WrongType
}

func NumericMatchUint64(op NumericOp, a *SexpUint64, b Sexp) (Sexp, error) {
	switch tb := b.(type) {
	case *SexpFloat:
		return NumericFloatDo(op, &SexpFloat{Val: float64(a.Val)}, tb), nil
	case *SexpInt:
		return NumericUint64Do(op, a, &SexpUint64{Val: uint64(tb.Val)}), nil
	case *SexpUint64:
		return NumericUint64Do(op, a, tb), nil
	case *SexpChar:
		return NumericUint64Do(op, a, &SexpUint64{Val: uint64(tb.Val)}), nil
	}
	return SexpNull, WrongType
}

func NumericMatchChar(op NumericOp, a *SexpChar, b Sexp) (Sexp, error) {
	var res Sexp
	switch tb := b.(type) {
	case *SexpFloat:
		res = NumericFloatDo(op, &SexpFloat{Val: float64(a.Val)}, tb)
	case *SexpInt:
		res = NumericIntDo(op, &SexpInt{Val: int64(a.Val)}, tb)
	case *SexpUint64:
		return NumericUint64Do(op, &SexpUint64{Val: uint64(a.Val)}, tb), nil
	case *SexpChar:
		res = NumericIntDo(op, &SexpInt{Val: int64(a.Val)}, &SexpInt{Val: int64(tb.Val)})
	default:
		return SexpNull, WrongType
	}
	switch tres := res.(type) {
	case *SexpFloat:
		return tres, nil
	case *SexpInt:
		return &SexpChar{Val: rune(tres.Val)}, nil
	}
	return SexpNull, errors.New("unexpected result")
}

func NumericDo(op NumericOp, a, b Sexp) (Sexp, error) {
	switch ta := a.(type) {
	case *SexpFloat:
		return NumericMatchFloat(op, ta, b)
	case *SexpInt:
		return NumericMatchInt(op, ta, b)
	case *SexpUint64:
		return NumericMatchUint64(op, ta, b)
	case *SexpChar:
		return NumericMatchChar(op, ta, b)
	}
	return SexpNull, WrongType
}
