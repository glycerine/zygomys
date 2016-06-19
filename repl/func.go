package zygo

import (
	"fmt"
)

func FuncBuilder(env *Glisp, name string,
	args []Sexp) (Sexp, error) {

	useName := name
	isMethod := false
	if name == "method" {
		isMethod = true
		useName = "method [p: (* StructName)]"
	}

	use := "use: (" + useName + " funcName [inputs:type ...] [returns:type ...])"

	n := len(args)
	if n < 1 {
		return SexpNull, fmt.Errorf("missing arguments. %s", use)
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
		if isMethod {
			ok := false
			symN, ok = args[1].(*SexpSymbol)
			if !ok {
				return SexpNull, fmt.Errorf("bad method name: symbol required after receiver array")
			}
			inputsLoc++
			returnsLoc++
			bodyLoc++
		} else {
			// anonymous function
			symN = env.GenSymbol("__anon")
			isAnon = true
			inputsLoc--
			returnsLoc--
			bodyLoc--
		}
	default:
		return SexpNull, fmt.Errorf("bad func name: symbol required")
	}
	Q("good: have func name '%v'", symN.name)
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
		return SexpNull, fmt.Errorf("bad func declaration '%v': "+
			"expected array of input declarations after the name. %s", funcName, use)
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

	Q("in func builder, args = ")
	for i := range args {
		Q("args[%v] = '%s'", i, args[i].SexpString(nil))
	}
	Q("in func builder, isAnon = %v", isAnon)
	Q("in func builder, inputs = %v", inputs.SexpString(nil))
	Q("in func builder, returns = %v", returns.SexpString(nil))
	Q("in func builder, body = %v", (&SexpArray{Val: body, Env: env}).SexpString(nil))

	inHash, err := GetFuncArgArray(inputs, env, "inputs")
	if err != nil {
		return SexpNull, fmt.Errorf("inputs array parsing error: %v", err)
	}
	Q("inHash = '%v'", inHash.SexpString(nil))

	retHash, err := GetFuncArgArray(returns, env, "returns")
	if err != nil {
		return SexpNull, fmt.Errorf("returns array parsing error: %v", err)
	}
	Q("retHash = '%v'", retHash.SexpString(nil))

	env.datastack.PushExpr(SexpNull)

	Q("FuncBuilder() about to call buildSexpFun")

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

	afsHelper := &AddFuncScopeHelper{}
	gen.AddInstruction(AddFuncScopeInstr{Name: "runtime " + gen.funcname, Helper: afsHelper})

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

	// minimal sanity check that we return the number of arguments
	// on the stack that are declared
	if len(body) == 0 {
		for range retHash.KeyOrder {
			gen.AddInstruction(PushInstr{expr: SexpNull})
		}
	}

	gen.AddInstruction(RemoveScopeInstr{})
	gen.AddInstruction(ReturnInstr{nil}) // nil is the error returned

	newfunc := GlispFunction(gen.instructions)
	sfun := gen.env.MakeFunction(gen.funcname, nargs,
		varargs, newfunc, orig)
	sfun.inputTypes = inHash
	sfun.returnTypes = retHash

	// tell the function scope where their function is, to
	// provide access to the captured-closure scopes at runtime.
	afsHelper.MyFunction = sfun

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

	if len(body) > 0 {
		invok.(*SexpFunction).hasBody = true
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

		//P("name = %#v", name)
		//P("typ  = %#v", typ)

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
					where, b.SexpString(nil))
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
					where, b.SexpString(nil))
			}
		}

		//P("here is env.ShowGlobalStack():")
		//env.ShowGlobalStack()

		//P("symN   = '%s'", symN.SexpString(nil))
		//P("symTyp = '%s'", symTyp.SexpString(nil))

		r, err, _ := env.LexicalLookupSymbol(symTyp, nil)
		if err != nil {
			return nil, fmt.Errorf("could not identify type %s: %v", symTyp.SexpString(nil), err)
		}
		switch rt := r.(type) {
		case *RegisteredType:
			// good, store it
			hash.HashSet(symN, rt)
		default:
			return nil, fmt.Errorf("'%s' is not a known type", symTyp.SexpString(nil))
		}
	}

	return hash, nil
}
