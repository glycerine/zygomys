package zygo

type SexpComment struct {
	Comment string
	Block   bool
}

func (p *SexpComment) SexpString(indent int) string {
	return p.Comment
}

func (p *SexpComment) Type() *RegisteredType {
	return GoStructRegistry.Registry["comment"]
}

// Filters return true to keep, false to drop.
type Filter func(x Sexp) bool

func RemoveCommentsFilter(x Sexp) bool {
	switch x.(type) {
	case *SexpComment:
		//P("RemoveCommentsFilter called on x= %T/val=%v. return false", x, x)
		return false
	default:
		//P("RemoveCommentsFilter called on x= %T/val=%v. return true", x, x)
		return true
	}
}

// detect SexpEnd values and return false on them to filter them out.
func RemoveEndsFilter(x Sexp) bool {
	switch n := x.(type) {
	case *SexpSentinel:
		if n.Val == SexpEnd.Val {
			return false
		}
	}
	return true
}

// detect SexpComma values and return false on them to filter them out.
func RemoveCommasFilter(x Sexp) bool {
	switch x.(type) {
	case *SexpComma:
		return false
	}
	return true
}

func (env *Glisp) FilterAny(x Sexp, f Filter) (filtered Sexp, keep bool) {
	switch ele := x.(type) {
	case *SexpArray:
		res := &SexpArray{Val: env.FilterArray(ele.Val, f), Typ: ele.Typ, IsFuncDeclTypeArray: ele.IsFuncDeclTypeArray, Env: env}
		return res, true
	case *SexpPair:
		return env.FilterList(ele, f), true
	case *SexpHash:
		return env.FilterHash(ele, f), true
	default:
		keep = f(x)
		if keep {
			return x, true
		}
		return SexpNull, false
	}
}

func (env *Glisp) FilterArray(x []Sexp, f Filter) []Sexp {
	//P("FilterArray: before: %d in size", len(x))
	//for i := range x {
	//P("x[i=%d] = %v", i, x[i].SexpString())
	//}
	res := []Sexp{}
	for i := range x {
		filtered, keep := env.FilterAny(x[i], f)
		if keep {
			res = append(res, filtered)
		}
	}
	//P("FilterArray: after: %d in size", len(res))
	//for i := range res {
	//P("x[i=%d] = %v", i, res[i].SexpString())
	//}
	return res
}

func (env *Glisp) FilterHash(h *SexpHash, f Filter) *SexpHash {
	// should not actually need this, since hashes
	// don't yet exist in parsed symbols. (they are
	// still lists).
	//P("in FilterHash")
	return h
}

func (env *Glisp) FilterList(h *SexpPair, f Filter) Sexp {
	//P("in FilterList")
	arr, err := ListToArray(h)
	res := []Sexp{}
	if err == NotAList {
		// don't filter pair lists
		return h
	}
	res = env.FilterArray(arr, f)
	return MakeList(res)
}

/*
func (env *Glisp) FilterDottedPair(h *SexpPair, f Filter) Sexp {

	res := &SexpPair{}
	ft, keepTail := env.FilterAny(h.Tail, f)
	if keepTail {
		res.Tail = ft
	}

	fh, keepHead := env.FilterAny(h.Head, f)
	if keepHead {
		res.Head = fh
	}
	switch {
	case keepHead && keepTail:
		return res
	case keepHead && !keepTail:
		return fh
	case !keepHead && keepTail:
		return ft
	}
	return SexpNull
}
*/
