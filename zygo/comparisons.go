package zygo

import (
	"bytes"
	"errors"
	"fmt"
	"math"
	"reflect"
)

func IsNaNFunction(env *Zlisp, name string, args []Sexp) (Sexp, error) {
	if len(args) != 1 {
		return SexpNull, WrongNargs
	}

	var err error
	a := args[0]
	if sel, isSel := a.(Selector); isSel {
		a, err = sel.RHS(env)
		if err != nil {
			return SexpNull, err
		}
	}

	switch at := a.(type) {
	case *SexpFloat:
		if math.IsNaN(at.Val) {
			return &SexpBool{Val: true}, nil
		}
	}
	return &SexpBool{Val: false}, nil
}

func signumFloat(f float64) int {
	if f > 0 {
		return 1
	}
	if f < 0 {
		return -1
	}
	return 0
}

func signumInt(i int64) int {
	if i > 0 {
		return 1
	}
	if i < 0 {
		return -1
	}
	return 0
}

func compareFloat(f *SexpFloat, expr Sexp) (int, error) {
	switch e := expr.(type) {
	case *SexpInt:
		return signumFloat(f.Val - float64(e.Val)), nil
	case *SexpFloat:
		nanCount := 0
		if math.IsNaN(f.Val) {
			nanCount++
		}
		if math.IsNaN(e.Val) {
			nanCount++
		}
		if nanCount > 0 {
			return 1 + nanCount, nil
		}
		return signumFloat(f.Val - e.Val), nil
	case *SexpChar:
		return signumFloat(f.Val - float64(e.Val)), nil
	}
	errmsg := fmt.Sprintf("err 91: cannot compare %T to %T", f, expr)
	return 0, errors.New(errmsg)
}

func compareInt(i *SexpInt, expr Sexp) (int, error) {
	switch e := expr.(type) {
	case *SexpInt:
		return signumInt(i.Val - e.Val), nil
	case *SexpFloat:
		return signumFloat(float64(i.Val) - e.Val), nil
	case *SexpChar:
		return signumInt(i.Val - int64(e.Val)), nil
	case *SexpReflect:
		r := reflect.Value(e.Val)
		ifa := r.Interface()
		switch z := ifa.(type) {
		case *int64:
			return signumInt(i.Val - *z), nil
		}
		P("compareInt(): ifa = %v/%T", ifa, ifa)
		P("compareInt(): r.Elem() = %v/%T", r.Elem(), r.Elem())
		P("compareInt(): r.Elem().Interface() = %v/%T", r.Elem().Interface(), r.Elem().Interface())
		P("compareInt(): r.Elem().Type() = %v/%T", r.Elem().Type(), r.Elem().Type())
		P("compareInt(): r.Elem().Type().Name() = %v/%T", r.Elem().Type().Name(), r.Elem().Type().Name())
	}
	errmsg := fmt.Sprintf("err 92: cannot compare %T to %T", i, expr)
	return 0, errors.New(errmsg)
}

func compareChar(c *SexpChar, expr Sexp) (int, error) {
	switch e := expr.(type) {
	case *SexpInt:
		return signumInt(int64(c.Val) - e.Val), nil
	case *SexpFloat:
		return signumFloat(float64(c.Val) - e.Val), nil
	case *SexpChar:
		return signumInt(int64(c.Val) - int64(e.Val)), nil
	}
	errmsg := fmt.Sprintf("err 93: cannot compare %T to %T", c, expr)
	return 0, errors.New(errmsg)
}

func compareString(s *SexpStr, expr Sexp) (int, error) {
	switch e := expr.(type) {
	case *SexpStr:
		return bytes.Compare([]byte(s.S), []byte(e.S)), nil
	case *SexpReflect:
		r := reflect.Value(e.Val)
		ifa := r.Interface()
		switch z := ifa.(type) {
		case *string:
			return bytes.Compare([]byte(s.S), []byte(*z)), nil
		}

	}
	errmsg := fmt.Sprintf("err 94: cannot compare %T to %T", s, expr)
	return 0, errors.New(errmsg)
}

func (env *Zlisp) compareSymbol(sym *SexpSymbol, expr Sexp) (int, error) {
	switch e := expr.(type) {
	case *SexpSymbol:
		return signumInt(int64(sym.number - e.number)), nil
	}
	errmsg := fmt.Sprintf("err 95: cannot compare %T to %T", sym, expr)
	return 0, errors.New(errmsg)
}

func (env *Zlisp) comparePair(a *SexpPair, b Sexp) (int, error) {
	var bp *SexpPair
	switch t := b.(type) {
	case *SexpPair:
		bp = t
	default:
		errmsg := fmt.Sprintf("err 96: cannot compare %T to %T", a, b)
		return 0, errors.New(errmsg)
	}
	res, err := env.Compare(a.Head, bp.Head)
	if err != nil {
		return 0, err
	}
	if res != 0 {
		return res, nil
	}
	return env.Compare(a.Tail, bp.Tail)
}

func (env *Zlisp) compareArray(a *SexpArray, b Sexp) (int, error) {
	var ba *SexpArray
	switch t := b.(type) {
	case *SexpArray:
		ba = t
	default:
		errmsg := fmt.Sprintf("err 97: cannot compare %T to %T", a, b)
		return 0, errors.New(errmsg)
	}
	var length int
	if len(a.Val) < len(ba.Val) {
		length = len(a.Val)
	} else {
		length = len(ba.Val)
	}

	for i := 0; i < length; i++ {
		res, err := env.Compare(a.Val[i], ba.Val[i])
		if err != nil {
			return 0, err
		}
		if res != 0 {
			return res, nil
		}
	}

	return signumInt(int64(len(a.Val) - len(ba.Val))), nil
}

func compareBool(a *SexpBool, b Sexp) (int, error) {
	var bb *SexpBool
	switch bt := b.(type) {
	case *SexpBool:
		bb = bt
	default:
		errmsg := fmt.Sprintf("err 98: cannot compare %T to %T", a, b)
		return 0, errors.New(errmsg)
	}

	// true > false
	if a.Val && bb.Val {
		return 0, nil
	}
	if a.Val {
		return 1, nil
	}
	if bb.Val {
		return -1, nil
	}
	return 0, nil
}

func comparePointers(a *SexpPointer, bs Sexp) (int, error) {
	var b *SexpPointer
	switch bt := bs.(type) {
	case *SexpPointer:
		b = bt
	default:
		return 0, fmt.Errorf("err 99: cannot compare %T to %T", a, bs)
	}

	if a.Target == b.Target {
		return 0, nil
	}
	return 1, nil
}

func (env *Zlisp) Compare(a Sexp, b Sexp) (int, error) {

	var err error
	if sel, isSel := a.(Selector); isSel {
		a, err = sel.RHS(env)
		if err != nil {
			return 0, err
		}
	}
	if sel, isSel := b.(Selector); isSel {
		b, err = sel.RHS(env)
		if err != nil {
			return 0, err
		}
	}

	switch at := a.(type) {
	case *SexpInt:
		return compareInt(at, b)
	case *SexpUint64:
		return compareUint64(at, b)
	case *SexpChar:
		return compareChar(at, b)
	case *SexpFloat:
		return compareFloat(at, b)
	case *SexpBool:
		return compareBool(at, b)
	case *SexpStr:
		return compareString(at, b)
	case *SexpSymbol:
		return env.compareSymbol(at, b)
	case *SexpPair:
		return env.comparePair(at, b)
	case *SexpArray:
		return env.compareArray(at, b)
	case *SexpHash:
		return compareHash(at, b)
	case *RegisteredType:
		return compareRegisteredTypes(at, b)
	case *SexpPointer:
		return comparePointers(at, b)
	case *SexpSentinel:
		if at == SexpNull && b == SexpNull {
			return 0, nil
		} else {
			return -1, nil
		}
	case *SexpReflect:
		r := reflect.Value(at.Val)
		ifa := r.Interface()
		//P("Compare(): ifa = %v/%t", ifa, ifa)
		//P("Compare(): r.Elem() = %v/%T", r.Elem(), r.Elem())
		switch z := ifa.(type) {
		case *int64:
			return compareInt(&SexpInt{Val: *z}, b)
		case *string:
			return compareString(&SexpStr{S: *z}, b)
		}

	}
	errmsg := fmt.Sprintf("err 100: cannot compare %T to %T", a, b)
	return 0, errors.New(errmsg)
}

// only compare uint64 to uint64
func compareUint64(i *SexpUint64, expr Sexp) (int, error) {
	switch e := expr.(type) {
	case *SexpUint64:
		return signumUint64(i.Val - e.Val), nil
	}
	errmsg := fmt.Sprintf("err 101: cannot compare %T to %T", i, expr)
	return 0, errors.New(errmsg)
}

func signumUint64(i uint64) int {
	if i > 0 {
		return 1
	}
	if i < 0 {
		return -1
	}
	return 0
}
