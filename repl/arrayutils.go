package zygo

import "fmt"

func MapArray(env *Glisp, fun *SexpFunction, arr *SexpArray) (Sexp, error) {
	result := make([]Sexp, len(arr.Val))
	var err error

	var firstTyp *RegisteredType
	for i := range arr.Val {
		result[i], err = env.Apply(fun, arr.Val[i:i+1])
		if err != nil {
			return &SexpArray{Val: result, Typ: firstTyp}, err
		}
		if firstTyp == nil {
			firstTyp = result[i].Type()
		}
	}

	return &SexpArray{Val: result, Typ: firstTyp}, nil
}

func ConcatArray(arr *SexpArray, rest []Sexp) (Sexp, error) {
	if arr == nil {
		return SexpNull, fmt.Errorf("ConcatArray called with nil arr")
	}
	var res SexpArray
	res.Val = arr.Val
	for i, x := range rest {
		switch t := x.(type) {
		case *SexpArray:
			res.Val = append(res.Val, t.Val...)
		default:
			return &res, fmt.Errorf("ConcatArray error: %d-th argument "+
				"(0-based) is not an array", i)
		}
	}
	return &res, nil
}

// (arrayidx ar [0 1])
func ArrayIndexFunction(env *Glisp, name string, args []Sexp) (Sexp, error) {
	//P("in ArrayIndexFunction, args = '%#v'", args)
	narg := len(args)
	if narg != 2 {
		return SexpNull, WrongNargs
	}

	var err error
	args, err = env.ResolveDotSym(args)
	if err != nil {
		return SexpNull, err
	}

	var ar *SexpArray
	switch ar2 := args[0].(type) {
	case *SexpSelector:
		x, err := ar2.RHS(env)
		if err != nil {
			return SexpNull, err
		}
		switch xArr := x.(type) {
		case *SexpArray:
			ar = xArr
		default:
			return SexpNull, fmt.Errorf("bad (arrayidx ar index) call: ar as arrayidx, but that did not resolve to an array, instead '%s'/type %T", x.SexpString(0), x)
		}
	case *SexpArray:
		ar = ar2
	default:
		return SexpNull, fmt.Errorf("bad (arrayidx ar index) call: ar was not an array, instead '%s'/type %T",
			args[0].SexpString(0), args[0])
	}

	var idx *SexpArray
	switch idx2 := args[1].(type) {
	case *SexpArray:
		idx = idx2
	default:
		return SexpNull, fmt.Errorf("bad (arrayidx ar index) call: index was not an array, instead '%s'/type %T",
			args[1].SexpString(0), args[1])
	}

	ret := SexpSelector{}
	ret.Select = idx
	ret.Container = ar

	return &ret, nil
}

// IndexBy subsets one array (possibly multidimensional) by another.
// e.g. if arr is [a b c] and idx is [0], we'll return a.
func (arr *SexpArray) IndexBy(idx *SexpArray) (Sexp, error) {
	nIdx := len(idx.Val)
	nTarget := arr.NumDim()

	if nIdx > nTarget {
		return SexpNull, fmt.Errorf("bad (arrayidx ar index) call: index requested %d dimensions, only have %d",
			nIdx, nTarget)
	}

	if len(idx.Val) == 0 {
		return SexpNull, fmt.Errorf("bad (arrayidx ar index) call: no index supplied")
	}
	if len(idx.Val) != 1 {
		return SexpNull, fmt.Errorf("bad (arrayidx ar index) call: we only support a single index value atm")
	}

	i := 0
	myInt, isInt := idx.Val[i].(*SexpInt)
	if !isInt {
		return SexpNull, fmt.Errorf("bad (arrayidx ar index) call: index with non-integer '%v'",
			idx.Val[i].SexpString(0))
	}
	k := myInt.Val
	pos := k % int64(len(arr.Val))
	if k < 0 {
		mk := -k
		mod := mk % int64(len(arr.Val))
		pos = int64(len(arr.Val)) - mod
	}
	//P("return pos %v", pos)
	return arr.Val[pos], nil
}

func (arr *SexpArray) NumDim() int {
	return 1
}

// SexpSelector: select a subset of an array:
// can be multidimensional index/slice
// and hence know its container and its position(s),
// and thus be able to read and write that position as
// need be.
type SexpSelector struct {
	Select    *SexpArray
	Container *SexpArray
}

func (si *SexpSelector) SexpString(indent int) string {
	rhs, err := si.RHS(nil)
	if err != nil {
		return fmt.Sprintf("(arraySelector %v %v)", si.Container.SexpString(indent), si.Select.SexpString(indent))
	}
	return fmt.Sprintf("%v /*(arraySelector %v %v)*/", rhs.SexpString(indent), si.Container.SexpString(indent), si.Select.SexpString(indent))
}

// Type returns the type of the value.
func (si *SexpSelector) Type() *RegisteredType {
	return GoStructRegistry.Lookup("arraySelector")
}

// RHS applies the selector to the contain and returns
// the value obtained.
func (x *SexpSelector) RHS(env *Glisp) (Sexp, error) {
	if len(x.Select.Val) != 1 {
		return SexpNull, fmt.Errorf("SexpSelector: only " +
			"size 1 selectors implemented")
	}
	var i int64
	switch asInt := x.Select.Val[0].(type) {
	case *SexpInt:
		i = asInt.Val
	default:
		return SexpNull, fmt.Errorf("SexpSelector: int "+
			"selector required; we saw %T", x.Select.Val[0])
	}
	if i < 0 {
		return SexpNull, fmt.Errorf("SexpSelector: negative "+
			"indexes not supported; we saw %v", i)
	}
	if i >= int64(len(x.Container.Val)) {
		return SexpNull, fmt.Errorf("SexpSelector: index "+
			"%v is out-of-bounds; length is %v", i, len(x.Container.Val))
	}
	return x.Container.Val[i], nil
}

// HasRHS structs have a RHS (right-hand-side)
// method that can be used to dereference the pointer-
// like object, yielding a value suitable for the
// right-hand-side of an assignment statement.
type HasRHS interface {
	RHS(env *Glisp) (Sexp, error)
}

func (x *SexpSelector) AssignToSelection(rhs Sexp) error {
	_, err := x.RHS(nil) // check for errors
	if err != nil {
		return err
	}
	x.Container.Val[x.Select.Val[0].(*SexpInt).Val] = rhs
	return nil
}
