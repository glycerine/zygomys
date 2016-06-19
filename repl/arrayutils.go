package zygo

import "fmt"

func MapArray(env *Glisp, fun *SexpFunction, arr *SexpArray) (Sexp, error) {
	result := make([]Sexp, len(arr.Val))
	var err error

	var firstTyp *RegisteredType
	for i := range arr.Val {
		result[i], err = env.Apply(fun, arr.Val[i:i+1])
		if err != nil {
			return &SexpArray{Val: result, Typ: firstTyp, Env: env}, err
		}
		if firstTyp == nil {
			firstTyp = result[i].Type()
		}
	}

	return &SexpArray{Val: result, Typ: firstTyp, Env: env}, nil
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
	Q("in ArrayIndexFunction, args = '%#v'", args)
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
	case *SexpArraySelector:
		x, err := ar2.RHS(env)
		if err != nil {
			return SexpNull, err
		}
		switch xArr := x.(type) {
		case *SexpArray:
			ar = xArr
		case *SexpHash:
			return HashIndexFunction(env, name, []Sexp{xArr, args[1]})
		default:
			return SexpNull, fmt.Errorf("bad (arrayidx ar index) call: ar as arrayidx, but that did not resolve to an array, instead '%s'/type %T", x.SexpString(nil), x)
		}
	case *SexpArray:
		ar = ar2
	case *SexpHash:
		return HashIndexFunction(env, name, args)
	case *SexpHashSelector:
		Q("ArrayIndexFunction sees args[0] is a hashSelector")
		return HashIndexFunction(env, name, args)
	default:
		return SexpNull, fmt.Errorf("bad (arrayidx ar index) call: ar was not an array, instead '%s'/type %T",
			args[0].SexpString(nil), args[0])
	}

	var idx *SexpArray
	switch idx2 := args[1].(type) {
	case *SexpArray:
		idx = idx2
	default:
		return SexpNull, fmt.Errorf("bad (arrayidx ar index) call: index was not an array, instead '%s'/type %T",
			args[1].SexpString(nil), args[1])
	}

	ret := SexpArraySelector{
		Select:    idx,
		Container: ar,
	}
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
			idx.Val[i].SexpString(nil))
	}
	k := myInt.Val
	pos := k % int64(len(arr.Val))
	if k < 0 {
		mk := -k
		mod := mk % int64(len(arr.Val))
		pos = int64(len(arr.Val)) - mod
	}
	//Q("return pos %v", pos)
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
type SexpArraySelector struct {
	Select    *SexpArray
	Container *SexpArray
}

func (si *SexpArraySelector) SexpString(ps *PrintState) string {
	Q("in SexpArraySelector.SexpString(), si.Container.Env = %p", si.Container.Env)
	rhs, err := si.RHS(si.Container.Env)
	if err != nil {
		return fmt.Sprintf("(arraySelector %v %v)", si.Container.SexpString(ps), si.Select.SexpString(ps))
	}

	Q("in SexpArraySelector.SexpString(), rhs = %v", rhs.SexpString(ps))
	Q("in SexpArraySelector.SexpString(), si.Container = %v", si.Container.SexpString(ps))
	Q("in SexpArraySelector.SexpString(), si.Select = %v", si.Select.SexpString(ps))

	return fmt.Sprintf("%v /*(arraySelector %v %v)*/", rhs.SexpString(ps), si.Container.SexpString(ps), si.Select.SexpString(ps))
}

// Type returns the type of the value.
func (si *SexpArraySelector) Type() *RegisteredType {
	return GoStructRegistry.Lookup("arraySelector")
}

// RHS applies the selector to the contain and returns
// the value obtained.
func (x *SexpArraySelector) RHS(env *Glisp) (Sexp, error) {
	if len(x.Select.Val) != 1 {
		return SexpNull, fmt.Errorf("SexpArraySelector: only " +
			"size 1 selectors implemented")
	}
	var i int64
	switch asInt := x.Select.Val[0].(type) {
	case *SexpInt:
		i = asInt.Val
	default:
		return SexpNull, fmt.Errorf("SexpArraySelector: int "+
			"selector required; we saw %T", x.Select.Val[0])
	}
	if i < 0 {
		return SexpNull, fmt.Errorf("SexpArraySelector: negative "+
			"indexes not supported; we saw %v", i)
	}
	if i >= int64(len(x.Container.Val)) {
		return SexpNull, fmt.Errorf("SexpArraySelector: index "+
			"%v is out-of-bounds; length is %v", i, len(x.Container.Val))
	}
	ret := x.Container.Val[i]
	Q("arraySelector returning ret = %#v", ret)
	return ret, nil
}

// Selector stores indexing information that isn't
// yet materialized for getting or setting.
//
type Selector interface {
	// RHS (right-hand-side) is used to dereference
	// the pointer-like Selector, yielding a value suitable for the
	// right-hand-side of an assignment statement.
	//
	RHS(env *Glisp) (Sexp, error)

	// AssignToSelection sets the selection to rhs
	// The selected elements are the left-hand-side of the
	// assignment *lhs = rhs
	AssignToSelection(env *Glisp, rhs Sexp) error
}

func (x *SexpArraySelector) AssignToSelection(env *Glisp, rhs Sexp) error {
	_, err := x.RHS(x.Container.Env) // check for errors
	if err != nil {
		return err
	}
	x.Container.Val[x.Select.Val[0].(*SexpInt).Val] = rhs
	return nil
}

func (env *Glisp) NewSexpArray(arr []Sexp) *SexpArray {
	return &SexpArray{Val: arr, Env: env}
}
