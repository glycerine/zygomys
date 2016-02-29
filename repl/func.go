package zygo

import (
	"fmt"
)

func FuncBuilder(env *Glisp, name string,
	args []Sexp) (Sexp, error) {

	use := "use: (func func-name [inputs:type ...] [returns:type ...])"

	n := len(args)
	if n < 1 {
		return SexpNull, fmt.Errorf("func definition missing arguments. %s", use)
	}

	inputsLoc := 1
	returnsLoc := 2
	bodyLoc := 3
	isAnon := false

	var symN *SexpSymbol
	switch b := args[0].(type) {
	case *SexpSymbol:
		symN = b
	case *SexpPair:
		sy, isQuo := isQuotedSymbol(b)
		if isQuo {
			symN = sy.(*SexpSymbol)
		} else {
			return SexpNull, fmt.Errorf("bad func name: symbol required")
		}
	case *SexpArray:
		// anonymous function
		symN = env.GenSymbol("__anon_func_")
		isAnon = true
		inputsLoc--
		returnsLoc--
		bodyLoc--
	default:
		return SexpNull, fmt.Errorf("bad func name: symbol required")
	}
	P("good: have func name '%v'", symN.name)
	funcName := symN.name

	if n < inputsLoc+1 {
		return SexpNull, fmt.Errorf("func [inputs] array is missing. %s", use)
	}

	if n < returnsLoc+1 {
		return SexpNull, fmt.Errorf("func [returns] array is missing. %s", use)
	}

	var inputs *SexpArray
	switch ar := args[inputsLoc].(type) {
	default:
		return SexpNull, fmt.Errorf("bad func declaration '%v': second argument "+
			"must be a array of input declarations. %s", funcName, use)
	case *SexpArray:
		inputs = ar
	}

	var returns *SexpArray
	switch ar := args[returnsLoc].(type) {
	default:
		return SexpNull, fmt.Errorf("bad func declaration '%v': third argument "+
			"must be a array of return declarations. %s", funcName, use)
	case *SexpArray:
		returns = ar
	}

	body := args[bodyLoc:]

	P("in func builder, args = ")
	for i := range args {
		P("args[%v] = '%s'", i, args[i].SexpString())
	}
	P("in func builder, isAnon = %v", isAnon)
	P("in func builder, inputs = %v", inputs.SexpString())
	P("in func builder, returns = %v", returns.SexpString())
	P("in func builder, body = %v", (&SexpArray{Val: body}).SexpString())
	env.datastack.PushExpr(SexpNull)

	/*
		err := env.LexicalBindSymbol(symN, rt)
		if err != nil {
			return SexpNull, fmt.Errorf("late: struct builder could not bind symbol '%s': '%v'",
				structName, err)
		}
		P("good: bound symbol '%s' to new func", symN.SexpString())
	*/
	return SexpNull, nil
}
