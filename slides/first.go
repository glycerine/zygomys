package zygo

// START OMIT
func FirstFunction(env *Glisp, name string, args []Sexp) (Sexp, error) {
	if len(args) != 1 {
		return SexpNull, WrongNargs
	}
	switch expr := args[0].(type) {
	case *SexpPair:
		return expr.Head, nil
	case *SexpArray:
		if len(expr.Val) > 0 {
			return expr.Val[0], nil
		}
		return SexpNull, fmt.Errorf("first called on empty array")
	}
	return SexpNull, WrongType
}

// END OMIT
