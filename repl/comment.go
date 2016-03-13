package zygo

type SexpComment struct {
	Comment string
	Block   bool
}

func (p *SexpComment) SexpString() string {
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

func (env *Glisp) FilterAny(x Sexp, f Filter) (filtered Sexp, keep bool) {
	switch ele := x.(type) {
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
	//P("in FilterHash")
	return h
}

func (env *Glisp) FilterList(h *SexpPair, f Filter) Sexp {
	//P("in FilterList")
	arr, err := ListToArray(h)
	panicOn(err)
	res := env.FilterArray(arr, f)
	return MakeList(res)
}
