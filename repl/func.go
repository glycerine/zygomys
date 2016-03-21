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
		symN = env.GenSymbol("__anon")
		isAnon = true
		inputsLoc--
		returnsLoc--
		bodyLoc--
	default:
		return SexpNull, fmt.Errorf("bad func name: symbol required")
	}
	P("good: have func name '%v'", symN.name)
	funcName := symN.name

	builtin, builtTyp := env.IsBuiltinSym(symN)
	if builtin {
		return SexpNull,
			fmt.Errorf("already have %s '%s', refusing to overwrite with defn",
				builtTyp, symN.name)
	}

	if env.HasMacro(symN) {
		return SexpNull, fmt.Errorf("Already have macro named '%s': refusing"+
			" to define function of same name.", symN.name)
	}

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
		inputs.IsFuncDeclTypeArray = true
	}

	var returns *SexpArray
	switch ar := args[returnsLoc].(type) {
	default:
		return SexpNull, fmt.Errorf("bad func declaration '%v': third argument "+
			"must be a array of return declarations. %s", funcName, use)
	case *SexpArray:
		returns = ar
		returns.IsFuncDeclTypeArray = true
	}

	body := args[bodyLoc:]

	P("in func builder, args = ")
	for i := range args {
		P("args[%v] = '%s'", i, args[i].SexpString(0))
	}
	P("in func builder, isAnon = %v", isAnon)
	P("in func builder, inputs = %v", inputs.SexpString(0))
	P("in func builder, returns = %v", returns.SexpString(0))
	P("in func builder, body = %v", (&SexpArray{Val: body, Env: env}).SexpString(0))

	inHash, err := GetFuncArgArray(inputs, env, "inputs")
	if err != nil {
		return SexpNull, fmt.Errorf("inputs array parsing error: %v", err)
	}
	P("inHash = '%v'", inHash.SexpString(0))

	retHash, err := GetFuncArgArray(returns, env, "returns")
	if err != nil {
		return SexpNull, fmt.Errorf("returns array parsing error: %v", err)
	}
	P("retHash = '%v'", retHash.SexpString(0))

	env.datastack.PushExpr(SexpNull)

	P("FuncBuilder() about to call buildSexpFun")

	// ===================================
	// ===================================
	//
	// from buildSexpFun, adapted
	//
	// todo: type checking the inputs and handling the returns as well
	//
	// ===================================
	// ===================================

	//	sfun, err := buildSexpFun(env, symN.name, funcargs, body, orig)

	//orig := &SexpArray{Val: args, Env: env}
	origa := []Sexp{env.MakeSymbol("func")}
	origa = append(origa, args...)
	orig := MakeList(origa)

	funcargs := inHash.KeyOrder

	gen := NewGenerator(env)
	gen.Tail = true

	gen.funcname = funcName

	gen.AddInstruction(AddFuncScopeInstr{Name: "runtime " + gen.funcname})

	argsyms := make([]*SexpSymbol, len(funcargs))

	// copy
	for i := range funcargs {
		argsyms[i] = funcargs[i].(*SexpSymbol)
	}

	varargs := false
	nargs := len(funcargs)

	if len(argsyms) >= 2 && argsyms[len(argsyms)-2].name == "&" {
		argsyms[len(argsyms)-2] = argsyms[len(argsyms)-1]
		argsyms = argsyms[0 : len(argsyms)-1]
		varargs = true
		nargs = len(argsyms) - 1
	}

	VPrintf("\n in buildSexpFun(): DumpFunction just before %v args go onto stack\n",
		len(argsyms))
	if Working {
		DumpFunction(GlispFunction(gen.instructions), -1)
	}
	for i := len(argsyms) - 1; i >= 0; i-- {
		gen.AddInstruction(PopStackPutEnvInstr{argsyms[i]})
	}
	err = gen.GenerateBegin(body)
	if err != nil {
		return MissingFunction, err
	}

	gen.AddInstruction(RemoveScopeInstr{})
	gen.AddInstruction(ReturnInstr{nil})

	newfunc := GlispFunction(gen.instructions)
	sfun := gen.env.MakeFunction(gen.funcname, nargs,
		varargs, newfunc, orig)
	sfun.inputTypes = inHash
	sfun.returnTypes = retHash

	clos := CreateClosureInstr{sfun}
	notePc := env.pc
	clos.Execute(env)

	invok, err := env.datastack.PopExpr()
	panicOn(err) // we just pushed in the clos.Execute(), so this should always be err == nil
	env.pc = notePc
	err = env.LexicalBindSymbol(symN, invok)
	if err != nil {
		return SexpNull, fmt.Errorf("internal error: could not bind symN:'%s' into env: %v", symN.name, err)
	}

	return invok, nil
}

func GetFuncArgArray(arr *SexpArray, env *Glisp, where string) (*SexpHash, error) {
	ar := arr.Val
	n := len(ar)
	hash, err := MakeHash([]Sexp{}, "hash", env)
	panicOn(err)
	if n == 0 {
		return hash, nil
	}
	if n%2 != 0 {
		return nil, fmt.Errorf("func definintion's %s array must have an even number of elements (each name:type pair counts as two)", where)
	}

	for i := 0; i < n; i += 2 {
		name := ar[i]
		typ := ar[i+1]

		P("name = %#v", name)
		P("typ  = %#v", typ)

		var symN *SexpSymbol
		switch b := name.(type) {
		case *SexpSymbol:
			symN = b
		case *SexpPair:
			sy, isQuo := isQuotedSymbol(b)
			if isQuo {
				symN = sy.(*SexpSymbol)
			} else {
				return nil, fmt.Errorf("bad formal parameter name: symbol required in %s array, not a symbol: '%s'",
					where, b.SexpString(0))
			}
		}

		var symTyp *SexpSymbol
		switch b := typ.(type) {
		case *SexpSymbol:
			symTyp = b
		case *SexpPair:
			sy, isQuo := isQuotedSymbol(b)
			if isQuo {
				symTyp = sy.(*SexpSymbol)
			} else {
				return nil, fmt.Errorf("bad formal parameter type: type required in %s array, but found '%s'",
					where, b.SexpString(0))
			}
		}

		P("symN   = '%s'", symN.SexpString(0))
		P("symTyp = '%s'", symTyp.SexpString(0))

		r, err, _ := env.LexicalLookupSymbol(symTyp, false)
		if err != nil {
			return nil, fmt.Errorf("could not identify type %s: %v", symTyp.SexpString(0), err)
		}
		switch rt := r.(type) {
		case *RegisteredType:
			// good, store it
			hash.HashSet(symN, rt)
		default:
			return nil, fmt.Errorf("'%s' is not a known type", symTyp.SexpString(0))
		}
	}

	return hash, nil
}
