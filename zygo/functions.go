package zygo

import (
	"bytes"
	"fmt"
	"os"
	"reflect"
	"runtime"
	"unicode"
)

var WrongNargs error = fmt.Errorf("wrong number of arguments")

type ZlispFunction []Instruction
type ZlispUserFunction func(*Zlisp, string, []Sexp) (Sexp, error)

func CompareFunction(name string) ZlispUserFunction {
	return func(env *Zlisp, _ string, args []Sexp) (Sexp, error) {
		if len(args) != 2 {
			return SexpNull, WrongNargs
		}

		res, err := env.Compare(args[0], args[1])
		if err != nil {
			return SexpNull, err
		}

		if res > 1 {
			//fmt.Printf("CompareFunction, res = %v\n", res)
			// 2 => one NaN found
			// 3 => two NaN found
			// NaN != NaN needs to return true.
			// NaN != 3.0 needs to return true.
			if name == "!=" {
				return &SexpBool{Val: true}, nil
			}
			return &SexpBool{Val: false}, nil
		}

		cond := false
		switch name {
		case "<":
			cond = res < 0
		case ">":
			cond = res > 0
		case "<=":
			cond = res <= 0
		case ">=":
			cond = res >= 0
		case "==":
			cond = res == 0
		case "!=":
			cond = res != 0
		}

		return &SexpBool{Val: cond}, nil
	}
}

func BinaryIntFunction(name string) ZlispUserFunction {
	return func(env *Zlisp, _ string, args []Sexp) (Sexp, error) {
		if len(args) != 2 {
			return SexpNull, WrongNargs
		}

		var op IntegerOp
		switch name {
		case "sll":
			op = ShiftLeft
		case "sra":
			op = ShiftRightArith
		case "srl":
			op = ShiftRightLog
		case "mod":
			op = Modulo
		}

		return IntegerDo(op, args[0], args[1])
	}
}

func BitwiseFunction(name string) ZlispUserFunction {
	return func(env *Zlisp, _ string, args []Sexp) (Sexp, error) {
		if len(args) != 2 {
			return SexpNull, WrongNargs
		}

		var op IntegerOp
		switch name {
		case "bitAnd":
			op = BitAnd
		case "bitOr":
			op = BitOr
		case "bitXor":
			op = BitXor
		}

		accum := args[0]
		var err error

		for _, expr := range args[1:] {
			accum, err = IntegerDo(op, accum, expr)
			if err != nil {
				return SexpNull, err
			}
		}
		return accum, nil
	}
}

func ComplementFunction(env *Zlisp, name string, args []Sexp) (Sexp, error) {
	if len(args) != 1 {
		return SexpNull, WrongNargs
	}

	switch t := args[0].(type) {
	case *SexpInt:
		return &SexpInt{Val: ^t.Val}, nil
	case *SexpChar:
		return &SexpChar{Val: ^t.Val}, nil
	}

	return SexpNull, fmt.Errorf("Argument to bitNot should be integer")
}

func PointerOrNumericFunction(name string) ZlispUserFunction {
	return func(env *Zlisp, _ string, args []Sexp) (Sexp, error) {
		n := len(args)
		if n == 0 {
			return SexpNull, WrongNargs
		}
		if n >= 2 {
			return NumericFunction(name)(env, name, args)
		}
		return PointerToFunction(env, name, args)
	}
}

func NumericFunction(name string) ZlispUserFunction {
	return func(env *Zlisp, _ string, args []Sexp) (Sexp, error) {
		if len(args) < 1 {
			return SexpNull, WrongNargs
		}
		var err error
		args, err = env.SubstituteRHS(args)
		if err != nil {
			return SexpNull, err
		}

		accum := args[0]
		var op NumericOp
		switch name {
		case "+":
			op = Add
		case "-":
			op = Sub
		case "*":
			op = Mult
		case "/":
			op = Div
		case "**":
			op = Pow
		}

		for _, expr := range args[1:] {
			accum, err = NumericDo(op, accum, expr)
			if err != nil {
				return SexpNull, err
			}
		}
		return accum, nil
	}
}

func ConsFunction(env *Zlisp, name string, args []Sexp) (Sexp, error) {
	if len(args) != 2 {
		return SexpNull, WrongNargs
	}

	return Cons(args[0], args[1]), nil
}

func FirstFunction(env *Zlisp, name string, args []Sexp) (Sexp, error) {
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

func RestFunction(env *Zlisp, name string, args []Sexp) (Sexp, error) {
	if len(args) != 1 {
		return SexpNull, WrongNargs
	}

	switch expr := args[0].(type) {
	case *SexpPair:
		return expr.Tail, nil
	case *SexpArray:
		if len(expr.Val) == 0 {
			return expr, nil
		}
		return &SexpArray{Val: expr.Val[1:], Env: env, Typ: expr.Typ}, nil
	case *SexpSentinel:
		if expr == SexpNull {
			return SexpNull, nil
		}
	}

	return SexpNull, WrongType
}

func SecondFunction(env *Zlisp, name string, args []Sexp) (Sexp, error) {
	if len(args) != 1 {
		return SexpNull, WrongNargs
	}
	switch expr := args[0].(type) {
	case *SexpPair:
		tail := expr.Tail
		switch p := tail.(type) {
		case *SexpPair:
			return p.Head, nil
		}
		return SexpNull, fmt.Errorf("list too small for second")
	case *SexpArray:
		if len(expr.Val) >= 2 {
			return expr.Val[1], nil
		}
		return SexpNull, fmt.Errorf("array too small for second")
	}

	return SexpNull, WrongType
}

func ArrayAccessFunction(name string) ZlispUserFunction {
	return func(env *Zlisp, _ string, args []Sexp) (Sexp, error) {
		narg := len(args)
		if narg < 2 || narg > 3 {
			return SexpNull, WrongNargs
		}

		var arr *SexpArray
		switch t := args[0].(type) {
		case *SexpArray:
			arr = t
		default:
			return SexpNull, fmt.Errorf("First argument of aget must be array")
		}

		var i int
		switch t := args[1].(type) {
		case *SexpInt:
			i = int(t.Val)
		case *SexpChar:
			i = int(t.Val)
		default:
			// can we evaluate it?
			res, err := EvalFunction(env, "eval-aget-index", []Sexp{args[1]})
			if err != nil {
				return SexpNull, fmt.Errorf("error during eval of "+
					"array-access position argument: %s", err)
			}
			switch j := res.(type) {
			case *SexpInt:
				i = int(j.Val)
			default:
				return SexpNull, fmt.Errorf("Second argument of aget could not be evaluated to integer; got j = '%#v'/type = %T", j, j)
			}
		}

		switch name {
		case "hget":
			fallthrough
		case "aget":
			if i < 0 || i >= len(arr.Val) {
				// out of bounds -- do we have a default?
				if narg == 3 {
					return args[2], nil
				}
				return SexpNull, fmt.Errorf("Array index out of bounds")
			}
			return arr.Val[i], nil
		case "aset":
			if len(args) != 3 {
				return SexpNull, WrongNargs
			}
			if i < 0 || i >= len(arr.Val) {
				return SexpNull, fmt.Errorf("Array index out of bounds")
			}
			arr.Val[i] = args[2]
		}
		return SexpNull, nil
	}
}

func SgetFunction(env *Zlisp, name string, args []Sexp) (Sexp, error) {
	if len(args) != 2 {
		return SexpNull, WrongNargs
	}

	var str *SexpStr
	switch t := args[0].(type) {
	case *SexpStr:
		str = t
	default:
		return SexpNull, fmt.Errorf("First argument of sget must be string")
	}

	var i int
	switch t := args[1].(type) {
	case *SexpInt:
		i = int(t.Val)
	case *SexpChar:
		i = int(t.Val)
	default:
		return SexpNull, fmt.Errorf("Second argument of sget must be integer")
	}

	return &SexpChar{Val: rune(str.S[i])}, nil
}

func HashAccessFunction(name string) ZlispUserFunction {
	return func(env *Zlisp, _ string, args []Sexp) (Sexp, error) {
		if len(args) < 1 || len(args) > 3 {
			return SexpNull, WrongNargs
		}

		// handle *SexpSelector
		container := args[0]
		var err error
		if ptr, isPtrLike := container.(Selector); isPtrLike {
			container, err = ptr.RHS(env)
			if err != nil {
				return SexpNull, err
			}
		}

		var hash *SexpHash
		switch e := container.(type) {
		case *SexpHash:
			hash = e
		default:
			return SexpNull, fmt.Errorf("first argument to h* function must be hash")
		}

		switch name {
		case "hget":
			if len(args) == 3 {
				return hash.HashGetDefault(env, args[1], args[2])
			}
			return hash.HashGet(env, args[1])
		case "hset":
			if len(args) != 3 {
				return SexpNull, WrongNargs
			}
			err := hash.HashSet(args[1], args[2])
			return SexpNull, err
		case "hdel":
			if len(args) != 2 {
				return SexpNull, WrongNargs
			}
			err := hash.HashDelete(args[1])
			return SexpNull, err
		case "keys":
			if len(args) != 1 {
				return SexpNull, WrongNargs
			}
			keys := make([]Sexp, 0)
			n := len(hash.KeyOrder)
			arr := &SexpArray{Env: env}
			for i := 0; i < n; i++ {
				keys = append(keys, (hash.KeyOrder)[i])

				// try to get a .Typ value going too... from the first available.
				if arr.Typ == nil {
					arr.Typ = (hash.KeyOrder)[i].Type()
				}
			}
			arr.Val = keys
			return arr, nil
		case "hpair":
			if len(args) != 2 {
				return SexpNull, WrongNargs
			}
			switch posreq := args[1].(type) {
			case *SexpInt:
				pos := int(posreq.Val)
				if pos < 0 || pos >= len(hash.KeyOrder) {
					return SexpNull, fmt.Errorf("hpair position request %d out of bounds", pos)
				}
				return hash.HashPairi(pos)
			default:
				return SexpNull, fmt.Errorf("hpair position request must be an integer")
			}
		}

		return SexpNull, nil
	}
}

func HashColonFunction(env *Zlisp, name string, args []Sexp) (Sexp, error) {
	if len(args) < 2 || len(args) > 3 {
		return SexpNull, WrongNargs
	}

	var hash *SexpHash
	switch e := args[1].(type) {
	case *SexpHash:
		hash = e
	default:
		return SexpNull, fmt.Errorf("second argument of (:field hash) must be a hash")
	}

	if len(args) == 3 {
		return hash.HashGetDefault(env, args[0], args[2])
	}
	return hash.HashGet(env, args[0])
}

func SliceFunction(env *Zlisp, name string, args []Sexp) (Sexp, error) {
	if len(args) != 3 {
		return SexpNull, WrongNargs
	}

	var start int
	var end int
	switch t := args[1].(type) {
	case *SexpInt:
		start = int(t.Val)
	case *SexpChar:
		start = int(t.Val)
	default:
		return SexpNull, fmt.Errorf("Second argument of slice must be integer")
	}

	switch t := args[2].(type) {
	case *SexpInt:
		end = int(t.Val)
	case *SexpChar:
		end = int(t.Val)
	default:
		return SexpNull, fmt.Errorf("Third argument of slice must be integer")
	}

	switch t := args[0].(type) {
	case *SexpArray:
		return &SexpArray{Val: t.Val[start:end], Env: env, Typ: t.Typ}, nil
	case *SexpStr:
		return &SexpStr{S: t.S[start:end]}, nil
	}

	return SexpNull, fmt.Errorf("First argument of slice must be array or string")
}

func LenFunction(env *Zlisp, name string, args []Sexp) (Sexp, error) {
	if len(args) != 1 {
		return SexpNull, WrongNargs
	}

	var err error
	args, err = env.ResolveDotSym(args)
	if err != nil {
		return SexpNull, err
	}

	switch t := args[0].(type) {
	case *SexpSentinel:
		if t == SexpNull {
			return &SexpInt{}, nil
		}
		break
	case *SexpArray:
		return &SexpInt{Val: int64(len(t.Val))}, nil
	case *SexpStr:
		return &SexpInt{Val: int64(len(t.S))}, nil
	case *SexpHash:
		return &SexpInt{Val: int64(HashCountKeys(t))}, nil
	case *SexpPair:
		n, err := ListLen(t)
		return &SexpInt{Val: int64(n)}, err
	default:
		P("in LenFunction with args[0] of type %T", t)
	}
	return &SexpInt{}, fmt.Errorf("argument must be string, list, hash, or array")
}

func AppendFunction(name string) ZlispUserFunction {
	return func(env *Zlisp, _ string, args []Sexp) (Sexp, error) {
		if len(args) != 2 {
			return SexpNull, WrongNargs
		}

		switch t := args[0].(type) {
		case *SexpArray:
			switch name {
			case "append":
				return &SexpArray{Val: append(t.Val, args[1]), Env: env, Typ: t.Typ}, nil
			case "appendslice":
				switch sl := args[1].(type) {
				case *SexpArray:
					return &SexpArray{Val: append(t.Val, sl.Val...), Env: env, Typ: t.Typ}, nil
				default:
					return SexpNull, fmt.Errorf("Second argument of appendslice must be slice")
				}
			default:
				return SexpNull, fmt.Errorf("unrecognized append variant: '%s'", name)
			}
		case *SexpStr:
			return AppendStr(t, args[1])
		}

		return SexpNull, fmt.Errorf("First argument of append must be array or string")
	}
}

func ConcatFunction(env *Zlisp, name string, args []Sexp) (Sexp, error) {
	if len(args) < 1 {
		return SexpNull, WrongNargs
	}
	var err error
	args, err = env.ResolveDotSym(args)
	if err != nil {
		return SexpNull, err
	}

	switch t := args[0].(type) {
	case *SexpArray:
		return ConcatArray(t, args[1:])
	case *SexpStr:
		return ConcatStr(t, args[1:])
	case *SexpPair:
		n := len(args)
		switch {
		case n == 1:
			return t, nil
		default:
			return ConcatLists(t, args[1:])
		}
	}

	return SexpNull, fmt.Errorf("expected strings, lists or arrays")
}

func ReadFunction(env *Zlisp, name string, args []Sexp) (sx Sexp, err error) {
	if len(args) != 1 {
		return SexpNull, WrongNargs
	}
	str := ""
	switch t := args[0].(type) {
	case *SexpStr:
		str = t.S
	default:
		return SexpNull, WrongType
	}
	env.parser.ResetAddNewInput(bytes.NewBuffer([]byte(str)))
	//exp, err := env.parser.ParseExpression(0)
	// have to use the iter interface...once.
	for reply := range env.parser.ParsingIter() {
		err = reply.Err
		if len(reply.Expr) > 0 {
			sx = reply.Expr[0]
		}
		break
	}
	return
}

func OldEvalFunction(env *Zlisp, name string, args []Sexp) (Sexp, error) {
	if len(args) != 1 {
		return SexpNull, WrongNargs
	}
	P("EvalFunction() called, name = '%s'; args = %#v", name, args)
	newenv := env.Duplicate()
	err := newenv.LoadExpressions(args)
	if err != nil {
		return SexpNull, fmt.Errorf("failed to compile expression")
	}
	newenv.pc = 0
	return newenv.Run()
}

// EvalFunction: new version doesn't use a duplicated environment,
// allowing eval to create closures under the lexical scope and
// to allow proper scoping in a package.
func EvalFunction(env *Zlisp, name string, args []Sexp) (Sexp, error) {
	if len(args) < 1 {
		return SexpNull, WrongNargs
	}
	//P("EvalFunction() called, name = '%s'; args = %#v", name, (&SexpArray{Val: args}).SexpString(0))

	// Instead of LoadExpressions:
	args = env.FilterArray(args, RemoveCommentsFilter)
	args = env.FilterArray(args, RemoveEndsFilter)

	startingDataStackSize := env.datastack.Size()

	gen := NewGenerator(env)
	err := gen.GenerateBegin(args)
	if err != nil {
		return SexpNull, err
	}

	newfunc := ZlispFunction(gen.instructions)
	orig := &SexpArray{Val: args}
	sfun := env.MakeFunction("evalGeneratedFunction", 0, false, newfunc, orig)

	err = env.CallFunction(sfun, 0)
	if err != nil {
		return SexpNull, err
	}

	var resultSexp Sexp
	resultSexp, err = env.Run()
	if err != nil {
		return SexpNull, err
	}

	err = env.ReturnFromFunction()

	// some sanity checks
	if env.datastack.Size() > startingDataStackSize {
		/*
			xtra := env.datastack.Size() - startingDataStackSize
			panic(fmt.Sprintf("we've left some extra stuff (xtra = %v) on the datastack "+
				"during eval, don't be sloppy, fix it now! env.datastack.Size()=%v, startingDataStackSize = %v",
				xtra, env.datastack.Size(), startingDataStackSize))
			P("warning: truncating datastack back to startingDataStackSize %v", startingDataStackSize)
		*/
		env.datastack.TruncateToSize(startingDataStackSize)
	}
	if env.datastack.Size() < startingDataStackSize {
		P("about panic, since env.datastack.Size() < startingDataStackSize, here is env dump:")
		env.DumpEnvironment()
		panic(fmt.Sprintf("we've shrunk the datastack during eval, don't be sloppy, fix it now! env.datastack.Size()=%v. startingDataStackSize=%v", env.datastack.Size(), startingDataStackSize))
	}

	return resultSexp, err
}

func TypeQueryFunction(name string) ZlispUserFunction {
	return func(env *Zlisp, _ string, args []Sexp) (Sexp, error) {
		if len(args) != 1 {
			return SexpNull, WrongNargs
		}

		var result bool

		switch name {
		case "type?":
			return TypeOf(args[0]), nil
		case "list?":
			result = IsList(args[0])
		case "null?":
			result = (args[0] == SexpNull)
		case "array?":
			result = IsArray(args[0])
		case "number?":
			result = IsNumber(args[0])
		case "float?":
			result = IsFloat(args[0])
		case "int?":
			result = IsInt(args[0])
		case "char?":
			result = IsChar(args[0])
		case "symbol?":
			result = IsSymbol(args[0])
		case "string?":
			result = IsString(args[0])
		case "hash?":
			result = IsHash(args[0])
		case "zero?":
			result = IsZero(args[0])
		case "empty?":
			result = IsEmpty(args[0])
		case "func?":
			result = IsFunc(args[0])
		}

		return &SexpBool{Val: result}, nil
	}
}

func PrintFunction(name string) ZlispUserFunction {
	return func(env *Zlisp, _ string, args []Sexp) (Sexp, error) {
		if len(args) < 1 {
			return SexpNull, WrongNargs
		}

		var str string

		switch expr := args[0].(type) {
		case *SexpStr:
			str = expr.S
		default:
			str = expr.SexpString(nil)
		}

		switch name {
		case "println":
			fmt.Println(str)
		case "print":
			fmt.Print(str)
		case "printf", "sprintf":
			if len(args) == 1 && name == "printf" {
				fmt.Print(str)
			} else {
				ar := make([]interface{}, len(args)-1)
				for i := 0; i < len(ar); i++ {
					switch x := args[i+1].(type) {
					case *SexpInt:
						ar[i] = x.Val
					case *SexpBool:
						ar[i] = x.Val
					case *SexpFloat:
						ar[i] = x.Val
					case *SexpChar:
						ar[i] = x.Val
					case *SexpStr:
						ar[i] = x.S
					case *SexpTime:
						ar[i] = x.Tm.In(NYC)
					default:
						ar[i] = args[i+1]
					}
				}
				if name == "printf" {
					fmt.Printf(str, ar...)
				} else {
					// sprintf
					return &SexpStr{S: fmt.Sprintf(str, ar...)}, nil
				}
			}
		}

		return SexpNull, nil
	}
}

func NotFunction(env *Zlisp, name string, args []Sexp) (Sexp, error) {
	if len(args) != 1 {
		return SexpNull, WrongNargs
	}

	result := &SexpBool{Val: !IsTruthy(args[0])}
	return result, nil
}

func ApplyFunction(env *Zlisp, name string, args []Sexp) (Sexp, error) {
	if len(args) != 2 {
		return SexpNull, WrongNargs
	}
	var fun *SexpFunction
	var funargs []Sexp

	switch e := args[0].(type) {
	case *SexpFunction:
		fun = e
	default:
		return SexpNull, fmt.Errorf("first argument must be function")
	}

	switch e := args[1].(type) {
	case *SexpArray:
		funargs = e.Val
	case *SexpPair:
		var err error
		funargs, err = ListToArray(e)
		if err != nil {
			return SexpNull, err
		}
	default:
		return SexpNull, fmt.Errorf("second argument must be array or list")
	}

	return env.Apply(fun, funargs)
}

func MapFunction(env *Zlisp, name string, args []Sexp) (Sexp, error) {
	if len(args) != 2 {
		return SexpNull, WrongNargs
	}
	var fun *SexpFunction

	//VPrintf("\n debug Map: args = '%#v'\n", args)

	switch e := args[0].(type) {
	case *SexpFunction:
		fun = e
	default:
		return SexpNull, fmt.Errorf("first argument must be function, but we had %T / val = '%#v'", e, e)
	}

	switch e := args[1].(type) {
	case *SexpArray:
		return MapArray(env, fun, e)
	case *SexpPair:
		x, err := MapList(env, fun, e)
		return x, err
	default:
		return SexpNull, fmt.Errorf("second argument must be array or list; we saw %T / val = %#v", e, e)
	}
}

func MakeArrayFunction(env *Zlisp, name string, args []Sexp) (Sexp, error) {
	if len(args) < 1 {
		return SexpNull, WrongNargs
	}

	var size int
	switch e := args[0].(type) {
	case *SexpInt:
		size = int(e.Val)
	default:
		return SexpNull, fmt.Errorf("first argument must be integer")
	}

	var fill Sexp
	if len(args) == 2 {
		fill = args[1]
	} else {
		fill = SexpNull
	}

	arr := make([]Sexp, size)
	for i := range arr {
		arr[i] = fill
	}

	return env.NewSexpArray(arr), nil
}

func ConstructorFunction(name string) ZlispUserFunction {

	switch name {
	case "unbase64", "base64", "flipbase64":
		return MakeRawUnbase64(name)
	}

	return func(env *Zlisp, _ string, args []Sexp) (Sexp, error) {
		switch name {
		case "array":
			return env.NewSexpArray(args), nil
		case "list":
			return MakeList(args), nil
		case "hash":
			return MakeHash(args, "hash", env)
		case "raw":
			return MakeRaw(args)
		case "field":
			//Q("making hash for field")
			h, err := MakeHash(args, "field", env)
			if err != nil {
				return SexpNull, err
			}
			fld := (*SexpField)(h)
			//Q("hash for field is: '%v'", fld.SexpString(nil))
			return fld, nil
		case "struct":
			return MakeHash(args, "struct", env)
		case "msgmap":
			switch len(args) {
			case 0:
				return MakeHash(args, name, env)
			default:
				var arr []Sexp
				var err error
				if len(args) > 1 {
					arr, err = ListToArray(args[1])
					if err != nil {
						return SexpNull, fmt.Errorf("error converting "+
							"'%s' arguments to an array: '%v'", name, err)
					}
				} else {
					arr = args[1:]
				}
				switch nm := args[0].(type) {
				case *SexpStr:
					return MakeHash(arr, nm.S, env)
				case *SexpSymbol:
					return MakeHash(arr, nm.name, env)
				default:
					return MakeHash(arr, name, env)
				}
			}
		}
		return SexpNull, fmt.Errorf("invalid constructor")
	}
}

func SymnumFunction(env *Zlisp, name string, args []Sexp) (Sexp, error) {
	if len(args) != 1 {
		return SexpNull, WrongNargs
	}

	switch t := args[0].(type) {
	case *SexpSymbol:
		return &SexpInt{Val: int64(t.number)}, nil
	}
	return SexpNull, fmt.Errorf("argument must be symbol")
}

var MissingFunction = &SexpFunction{name: "__missing", user: true}

func (env *Zlisp) MakeFunction(name string, nargs int, varargs bool,
	fun ZlispFunction, orig Sexp) *SexpFunction {
	var sfun SexpFunction
	sfun.name = name
	sfun.user = false
	sfun.nargs = nargs
	sfun.varargs = varargs
	sfun.fun = fun
	sfun.orig = orig
	sfun.SetClosing(NewClosing(name, env)) // snapshot the create env as of now.
	return &sfun
}

func MakeUserFunction(name string, ufun ZlispUserFunction) *SexpFunction {
	var sfun SexpFunction
	sfun.name = name
	sfun.user = true
	sfun.userfun = ufun
	return &sfun
}

func MakeBuilderFunction(name string, ufun ZlispUserFunction) *SexpFunction {
	sfun := MakeUserFunction(name, ufun)
	sfun.isBuilder = true
	return sfun
}

// MergeFuncMap returns the union of the two given maps
func MergeFuncMap(funcs ...map[string]ZlispUserFunction) map[string]ZlispUserFunction {
	n := make(map[string]ZlispUserFunction)

	for _, f := range funcs {
		for k, v := range f {
			// disallow dups, avoiding possible security implications and confusion generally.
			if _, dup := n[k]; dup {
				panic(fmt.Sprintf(" duplicate function '%s' not allowed", k))
			}
			n[k] = v
		}
	}
	return n
}

// SandboxSafeFuncs returns all functions that are safe to run in a sandbox
func SandboxSafeFunctions() map[string]ZlispUserFunction {
	return MergeFuncMap(
		CoreFunctions(),
		StrFunctions(),
		EncodingFunctions(),
	)
}

// AllBuiltinFunctions returns all built in functions
func AllBuiltinFunctions() map[string]ZlispUserFunction {
	return MergeFuncMap(
		CoreFunctions(),
		StrFunctions(),
		EncodingFunctions(),
		SystemFunctions(),
		ReflectionFunctions(),
	)
}

// CoreFunctions returns all of the core logic
func CoreFunctions() map[string]ZlispUserFunction {
	return map[string]ZlispUserFunction{
		"echo":      SetEchoPrintFlag,
		"pretty":    SetPrettyPrintFlag,
		"<":         CompareFunction("<"),
		">":         CompareFunction(">"),
		"<=":        CompareFunction("<="),
		">=":        CompareFunction(">="),
		"==":        CompareFunction("=="),
		"!=":        CompareFunction("!="),
		"isnan":     IsNaNFunction("isnan"),
		"isNaN":     IsNaNFunction("isNaN"),
		"sll":       BinaryIntFunction("sll"),
		"sra":       BinaryIntFunction("sra"),
		"srl":       BinaryIntFunction("srl"),
		"mod":       BinaryIntFunction("mod"),
		"+":         NumericFunction("+"),
		"-":         NumericFunction("-"),
		"*":         PointerOrNumericFunction("*"),
		"**":        NumericFunction("**"),
		"/":         NumericFunction("/"),
		"bitAnd":    BitwiseFunction("bitAnd"),
		"bitOr":     BitwiseFunction("bitOr"),
		"bitXor":    BitwiseFunction("bitXor"),
		"bitNot":    ComplementFunction,
		"read":      ReadFunction,
		"cons":      ConsFunction,
		"first":     FirstFunction,
		"second":    SecondFunction,
		"rest":      RestFunction,
		"car":       FirstFunction,
		"cdr":       RestFunction,
		"type?":     TypeQueryFunction("type?"),
		"list?":     TypeQueryFunction("list?"),
		"null?":     TypeQueryFunction("null?"),
		"array?":    TypeQueryFunction("array?"),
		"hash?":     TypeQueryFunction("hash?"),
		"number?":   TypeQueryFunction("number?"),
		"int?":      TypeQueryFunction("int?"),
		"float?":    TypeQueryFunction("float?"),
		"char?":     TypeQueryFunction("char?"),
		"symbol?":   TypeQueryFunction("symbol?"),
		"string?":   TypeQueryFunction("string?"),
		"zero?":     TypeQueryFunction("zero?"),
		"empty?":    TypeQueryFunction("empty?"),
		"func?":     TypeQueryFunction("func?"),
		"not":       NotFunction,
		"apply":     ApplyFunction,
		"map":       MapFunction,
		"makeArray": MakeArrayFunction,
		"aget":      ArrayAccessFunction("aget"),
		"aset":      ArrayAccessFunction("aset"),
		"sget":      SgetFunction,
		"hget":      GenericAccessFunction, // handles arrays or hashes
		//":":          ColonAccessFunction,
		"hset":        HashAccessFunction("hset"),
		"hdel":        HashAccessFunction("hdel"),
		"keys":        HashAccessFunction("keys"),
		"hpair":       GenericHpairFunction,
		"slice":       SliceFunction,
		"len":         LenFunction,
		"append":      AppendFunction("append"),
		"appendslice": AppendFunction("appendslice"),
		"concat":      ConcatFunction,
		"field":       ConstructorFunction("field"),
		"struct":      ConstructorFunction("struct"),
		"array":       ConstructorFunction("array"),
		"list":        ConstructorFunction("list"),
		"hash":        ConstructorFunction("hash"),
		"raw":         ConstructorFunction("raw"),

		"raw64":      ConstructorFunction("raw64"),
		"unbase64":   ConstructorFunction("unbase64"),
		"base64":     ConstructorFunction("base64"),
		"flipbase64": ConstructorFunction("flipbase64"),
		"isbase64":   IsBase64Function,
		"copyraw":    CopyRawFunction,

		"str":       StringifyFunction,
		"->":        ThreadMapFunction,
		"flatten":   FlattenToWordsFunction,
		"quotelist": QuoteListFunction,
		"=":         AssignmentFunction,
		":=":        AssignmentFunction,
		"fieldls":   GoFieldListFunction,
		"defined?":  DefinedFunction,
		"stop":      StopFunction,
		"joinsym":   JoinSymFunction,
		"GOOS":      GOOSFunction,
		"&":         AddressOfFunction,
		"derefSet":  DerefFunction("derefSet"),
		"deref":     DerefFunction("deref"),
		".":         DotFunction,
		"arrayidx":  ArrayIndexFunction,
		"hashidx":   HashIndexFunction,
		"asUint64":  AsUint64Function,
	}
}

func StrFunctions() map[string]ZlispUserFunction {
	return map[string]ZlispUserFunction{
		"nsplit": SplitStringOnNewlinesFunction, "split": SplitStringFunction,
		"chomp":   StringUtilFunction("chomp"),
		"trim":    StringUtilFunction("trim"),
		"println": PrintFunction("println"),
		"print":   PrintFunction("print"),
		"printf":  PrintFunction("printf"),
		"sprintf": PrintFunction("sprintf"),
		"raw2str": RawToStringFunction,
		"str2sym": Str2SymFunction,
		"sym2str": Sym2StrFunction,
		"gensym":  GensymFunction,
		"symnum":  SymnumFunction,
		"json2":   SimpleJSONStringFunction,
	}

}

func EncodingFunctions() map[string]ZlispUserFunction {
	return map[string]ZlispUserFunction{
		"json":      JsonFunction("json"),
		"unjson":    JsonFunction("unjson"),
		"msgpack":   JsonFunction("msgpack"),
		"unmsgpack": JsonFunction("unmsgpack"),
		"gob":       GobEncodeFunction,
		"msgmap":    ConstructorFunction("msgmap"),
	}
}

func ReflectionFunctions() map[string]ZlispUserFunction {
	return map[string]ZlispUserFunction{
		"methodls":              GoMethodListFunction,
		"_method":               CallGoMethodFunction,
		"registerDemoFunctions": ScriptFacingRegisterDemoStructs,
	}
}

func SystemFunctions() map[string]ZlispUserFunction {
	return map[string]ZlispUserFunction{
		"source":    SourceFileFunction,
		"togo":      ToGoFunction,
		"fromgo":    FromGoFunction,
		"dump":      GoonDumpFunction,
		"slurpf":    SlurpfileFunction,
		"writef":    WriteToFileFunction("writef"),
		"save":      WriteToFileFunction("save"),
		"bload":     ReadGreenpackFromFileFunction,
		"bsave":     WriteShadowGreenpackToFileFunction("bsave"),
		"greenpack": WriteShadowGreenpackToFileFunction("greenpack"),
		"owritef":   WriteToFileFunction("owritef"),
		"system":    SystemFunction,
		"exit":      ExitFunction,
		"_closdump": DumpClosureEnvFunction,
		"rmsym":     RemoveSymFunction,
		"typelist":  TypeListFunction,
		"setenv":    GetEnvFunction("setenv"),
		"getenv":    GetEnvFunction("getenv"),
		// not done "_call":     CallZMethodOnRecordFunction,
	}
}

func ThreadMapFunction(env *Zlisp, name string, args []Sexp) (Sexp, error) {
	if len(args) < 2 {
		return SexpNull, WrongNargs
	}

	h, isHash := args[0].(*SexpHash)
	if !isHash {
		return SexpNull, fmt.Errorf("-> error: first argument must be a hash or defmap")
	}

	field, err := threadingHelper(env, h, args[1:])
	if err != nil {
		return SexpNull, err
	}

	return field, nil
}

func threadingHelper(env *Zlisp, hash *SexpHash, args []Sexp) (Sexp, error) {
	if len(args) == 0 {
		panic("should not recur without arguments")
	}
	field, err := hash.HashGet(env, args[0])
	if err != nil {
		return SexpNull, fmt.Errorf("-> error: field '%s' not found",
			args[0].SexpString(nil))
	}
	if len(args) > 1 {
		h, isHash := field.(*SexpHash)
		if !isHash {
			return SexpNull, fmt.Errorf("request for field '%s' was "+
				"not on a hash or defmap; instead type %T with value '%#v'",
				args[1].SexpString(nil), field, field)
		}
		return threadingHelper(env, h, args[1:])
	}
	return field, nil
}

func StringifyFunction(env *Zlisp, name string, args []Sexp) (Sexp, error) {
	if len(args) != 1 {
		return SexpNull, WrongNargs
	}

	return &SexpStr{S: args[0].SexpString(nil)}, nil
}

func Sym2StrFunction(env *Zlisp, name string, args []Sexp) (Sexp, error) {
	if len(args) != 1 {
		return SexpNull, WrongNargs
	}

	switch t := args[0].(type) {
	case *SexpSymbol:
		r := &SexpStr{S: t.name}
		return r, nil
	}
	return SexpNull, fmt.Errorf("argument must be symbol")
}

func Str2SymFunction(env *Zlisp, name string, args []Sexp) (Sexp, error) {
	if len(args) != 1 {
		return SexpNull, WrongNargs
	}

	switch t := args[0].(type) {
	case *SexpStr:
		return env.MakeSymbol(t.S), nil
	}
	return SexpNull, fmt.Errorf("argument must be string")
}

// the (json2 objectToDisplayAsJSON) function
func SimpleJSONStringFunction(env *Zlisp, name string, args []Sexp) (Sexp, error) {
	if len(args) != 1 {
		return SexpNull, WrongNargs
	}
	ps := NewPrintState()
	ps.PrintJSON = true
	s := args[0].SexpString(ps)
	return &SexpStr{S: s}, nil
}

func GensymFunction(env *Zlisp, name string, args []Sexp) (Sexp, error) {
	n := len(args)
	switch {
	case n == 0:
		return env.GenSymbol("__gensym"), nil
	case n == 1:
		switch t := args[0].(type) {
		case *SexpStr:
			return env.GenSymbol(t.S), nil
		}
		return SexpNull, fmt.Errorf("argument must be string")
	default:
		return SexpNull, WrongNargs
	}
}

func ExitFunction(env *Zlisp, name string, args []Sexp) (Sexp, error) {
	if len(args) != 1 {
		return SexpNull, WrongNargs
	}
	switch e := args[0].(type) {
	case *SexpInt:
		os.Exit(int(e.Val))
	}
	return SexpNull, fmt.Errorf("argument must be int (the exit code)")
}

// handles arrays or hashes
func GenericAccessFunction(env *Zlisp, name string, args []Sexp) (Sexp, error) {
	if len(args) < 1 || len(args) > 3 {
		return SexpNull, WrongNargs
	}

	// handle *SexpSelector
	container := args[0]
	var err error
	if ptr, isPtrLike := container.(Selector); isPtrLike {
		container, err = ptr.RHS(env)
		if err != nil {
			return SexpNull, err
		}
	}

	switch container.(type) {
	case *SexpHash:
		return HashAccessFunction(name)(env, name, args)
	case *SexpArray:
		return ArrayAccessFunction(name)(env, name, args)
	}
	return SexpNull, fmt.Errorf("first argument to hget function must be hash or array")
}

var stopErr error = fmt.Errorf("stop")

func StopFunction(env *Zlisp, name string, args []Sexp) (Sexp, error) {
	narg := len(args)
	if narg > 1 {
		return SexpNull, WrongNargs
	}

	if narg == 0 {
		return SexpNull, stopErr
	}

	switch s := args[0].(type) {
	case *SexpStr:
		return SexpNull, fmt.Errorf("%v", s.S)
	}
	return SexpNull, stopErr
}

// the assignment function, =
func AssignmentFunction(env *Zlisp, name string, args []Sexp) (Sexp, error) {
	//Q("\n AssignmentFunction called with name ='%s'. args='%s'\n", name, env.NewSexpArray(args).SexpString(nil))

	narg := len(args)
	if narg != 2 {
		return SexpNull, fmt.Errorf("assignment requires two arguments: a left-hand-side and a right-hand-side argument")
	}

	var sym *SexpSymbol
	switch s := args[0].(type) {
	case *SexpSymbol:
		sym = s
	case Selector:
		err := s.AssignToSelection(env, args[1])
		return args[1], err

	default:
		return SexpNull, fmt.Errorf("assignment needs left-hand-side"+
			" argument to be a symbol; we got %T", s)
	}

	if !sym.isDot {
		//Q("assignment sees LHS symbol but is not dot, binding '%s' to '%s'\n", sym.name, args[1].SexpString(nil))
		err := env.LexicalBindSymbol(sym, args[1])
		if err != nil {
			return SexpNull, err
		}
		return args[1], nil
	}

	//Q("assignment calling dotGetSetHelper()\n")
	return dotGetSetHelper(env, sym.name, &args[1])
}

func JoinSymFunction(env *Zlisp, name string, args []Sexp) (Sexp, error) {
	narg := len(args)
	if narg == 0 {
		return SexpNull, nil
	}

	j := ""

	for k := range args {
		switch a := args[k].(type) {
		case *SexpPair:
			arr, err := ListToArray(args[k])
			if err != nil {
				return SexpNull, fmt.Errorf("error converting "+
					"joinsym arguments to an array: '%v'", err)
			}
			s, err := joinSymHelper(arr)
			if err != nil {
				return SexpNull, err
			}
			j += s

		case *SexpSymbol:
			j = j + a.name
		case *SexpArray:
			s, err := joinSymHelper(a.Val)
			if err != nil {
				return SexpNull, err
			}
			j += s
		default:
			return SexpNull, fmt.Errorf("error cannot joinsym type '%T' / val = '%s'", a, a.SexpString(nil))
		}
	}

	return env.MakeSymbol(j), nil
}

func joinSymHelper(arr []Sexp) (string, error) {
	j := ""
	for i := 0; i < len(arr); i++ {
		switch s := arr[i].(type) {
		case *SexpSymbol:
			j = j + s.name

		default:
			return "", fmt.Errorf("not a symbol: '%s'",
				arr[i].SexpString(nil))
		}
	}
	return j, nil
}

// '(a b c) -> ('a 'b 'c)
func QuoteListFunction(env *Zlisp, name string, args []Sexp) (Sexp, error) {
	narg := len(args)
	if narg != 1 {
		return SexpNull, WrongNargs
	}

	pair, ok := args[0].(*SexpPair)
	if !ok {
		return SexpNull, fmt.Errorf("list required")
	}

	arr, err := ListToArray(pair)
	if err != nil {
		return SexpNull, fmt.Errorf("error converting "+
			"quotelist arguments to an array: '%v'", err)
	}

	arr2 := []Sexp{}
	for _, v := range arr {
		arr2 = append(arr2, MakeList([]Sexp{env.MakeSymbol("quote"), v}))
	}

	return MakeList(arr2), nil
}

// helper used by dotGetSetHelper and sub-calls to check for private
func errIfPrivate(pathPart string, pkg *Stack) error {
	noDot := stripAnyDotPrefix(pathPart)

	// references through a package must be Public
	if !unicode.IsUpper([]rune(noDot)[0]) {
		return fmt.Errorf("Cannot access private member '%s' of package '%s'",
			noDot, pkg.PackageName)
	}
	return nil
}

// if setVal is nil, only get and return the lookup.
// Otherwise set and return the value we set.
func dotGetSetHelper(env *Zlisp, name string, setVal *Sexp) (Sexp, error) {
	path := DotPartsRegex.FindAllString(name, -1)
	//P("\n in dotGetSetHelper(), name = '%s', path = '%#v', setVal = '%#v'\n", name, path, setVal)
	if len(path) == 0 {
		return SexpNull, fmt.Errorf("internal error: DotFunction" +
			" path had zero length")
	}

	var ret Sexp = SexpNull
	var err error
	lenpath := len(path)

	if lenpath == 1 && setVal != nil {
		// single path element set, bind it now.
		a := stripAnyDotPrefix(path[0])
		asym := env.MakeSymbol(a)

		// check conflict
		//Q("asym = %#v\n", asym)
		builtin, typ := env.IsBuiltinSym(asym)
		if builtin {
			return SexpNull, fmt.Errorf("'%s' is a %s, cannot assign to it with dot-symbol", asym.name, typ)
		}

		err := env.LexicalBindSymbol(asym, *setVal)
		if err != nil {
			return SexpNull, err
		}
		return *setVal, nil
	}

	// handle multiple paths that index into hashes after the
	// the first

	key := stripAnyDotPrefix(path[0])
	//Q("\n in dotGetSetHelper(), looking up '%s'\n", key)
	keySym := env.MakeSymbol(key)
	ret, err, _ = env.LexicalLookupSymbol(keySym, nil)
	if err != nil {
		//Q("\n in dotGetSetHelper(), '%s' not found\n", key)
		return SexpNull, err
	}
	if lenpath == 1 {
		// single path element get, return it.
		return ret, err
	}

	// INVAR: lenpath > 1

	// package or hash? check for package
	pkg, isStack := ret.(*Stack)
	if isStack && pkg.IsPackage {
		//P("found a package: '%s'", pkg.SexpString(nil))

		exp, err := pkg.nestedPathGetSet(env, path[1:], setVal)
		if err != nil {
			return SexpNull, err
		}
		return exp, nil
	}

	// at least .a.b if not a.b.c. etc: multiple elements,
	// where .b and after
	// will index into hashes (.a must refer to a hash);
	// proceed deeper into the hashes.

	var h *SexpHash
	switch x := ret.(type) {
	case *SexpHash:
		h = x
	case *SexpReflect:
		// at least allow reading, if we can.
		if setVal != nil {
			return SexpNull, fmt.Errorf("can't set on an SexpReflect: on request for "+
				"field '%s' in non-record (instead of type %T)",
				stripAnyDotPrefix(path[1]), ret)
		}
		//P("functions.go DEBUG! SexpReflect value h is type: '%v', '%T', kind: '%v'", x.Val.Type(), x.Val.Interface(), x.Val.Type().Kind())
		if x.Val.Type().Kind() == reflect.Struct {
			//P("We have a struct! path[1]='%v', path='%#v'", path[1], path)
			if len(path) >= 2 && len(path[1]) > 0 {
				fieldName := stripAnyDotPrefix(path[1])
				//P("We have a struct! with dot request for member '%s'", fieldName)
				fld := x.Val.FieldByName(fieldName)
				if reflect.DeepEqual(fld, reflect.Value{}) {
					return SexpNull, fmt.Errorf("no such field '%s'", fieldName)
				}
				// ex:  We got back fld='20' of type int, kind=int
				//P("We got back fld='%v' of type %v, kind=%v", fld, fld.Type(), fld.Type().Kind())
				return GoToSexp(fld.Interface(), env)
			}
		}
		return SexpNull, fmt.Errorf("SexpReflect is not a struct: cannot get "+
			"field '%s' in non-struct (instead of type %T)",
			stripAnyDotPrefix(path[1]), ret)
	default:
		return SexpNull, fmt.Errorf("not a record: cannot get "+
			"field '%s' in non-record (instead of type %T)",
			stripAnyDotPrefix(path[1]), ret)
	}
	// have hash: rest of path handled in hashutils.go in nestedPathGet()
	//Q("\n in dotGetSetHelper(), about to call nestedPathGetSet() with"+
	//	"dotpaths = path[i+1:]='%#v\n", path[1:])
	exp, err := h.nestedPathGetSet(env, path[1:], setVal)
	if err != nil {
		return SexpNull, err
	}
	return exp, nil
}

func RemoveSymFunction(env *Zlisp, name string, args []Sexp) (Sexp, error) {
	narg := len(args)
	if narg != 1 {
		return SexpNull, WrongNargs
	}

	sym, ok := args[0].(*SexpSymbol)
	if !ok {
		return SexpNull, fmt.Errorf("symbol required, but saw %T/%v", args[0], args[0].SexpString(nil))
	}

	err := env.linearstack.DeleteSymbolFromTopOfStackScope(sym)
	return SexpNull, err
}

func GOOSFunction(env *Zlisp, name string, args []Sexp) (Sexp, error) {
	narg := len(args)
	if narg != 0 {
		return SexpNull, WrongNargs
	}
	return &SexpStr{S: runtime.GOOS}, nil
}

// check is a symbol/string/value is defined
func DefinedFunction(env *Zlisp, name string, args []Sexp) (Sexp, error) {
	//P("in DefinedFunction, args = '%#v'", args)
	narg := len(args)
	if narg != 1 {
		return SexpNull, WrongNargs
	}

	var checkme string
	switch nm := args[0].(type) {
	case *SexpStr:
		checkme = nm.S
	case *SexpSymbol:
		checkme = nm.name
	case *SexpFunction:
		return &SexpBool{Val: true}, nil
	default:
		return &SexpBool{Val: false}, nil
	}

	_, err, _ := env.LexicalLookupSymbol(env.MakeSymbol(checkme), nil)
	if err != nil {
		return &SexpBool{Val: false}, nil
	}
	return &SexpBool{Val: true}, nil
}

func AddressOfFunction(env *Zlisp, name string, args []Sexp) (Sexp, error) {
	narg := len(args)
	if narg != 1 {
		return SexpNull, WrongNargs
	}

	return NewSexpPointer(args[0]), nil
}

func DerefFunction(name string) ZlispUserFunction {
	return func(env *Zlisp, _ string, args []Sexp) (result Sexp, err error) {
		result = SexpNull

		defer func() {
			e := recover()
			if e != nil {
				//Q("in recover() of DerefFunction, e = '%#v'", e)
				switch ve := e.(type) {
				case *reflect.ValueError:
					err = ve
				default:
					err = fmt.Errorf("unknown typecheck error during %s: %v", name, ve)
				}
			}
		}()

		narg := len(args)
		if narg != 1 && narg != 2 {
			return SexpNull, WrongNargs
		}
		var ptr *SexpPointer
		switch e := args[0].(type) {
		case *SexpPointer:
			ptr = e
		case *SexpReflect:
			ptr = NewSexpPointer(e)
		default:
			return SexpNull, fmt.Errorf("%s only operates on pointers (*SexpPointer); we saw %T instead", name, e)
		}

		switch name {
		case "deref":
			if narg != 1 {
				return SexpNull, WrongNargs
			}
			return ptr.Target, nil

		case "derefSet":
			if narg != 2 {
				return SexpNull, WrongNargs
			}

			// delegate as much as we can to the Go type system
			// and reflection
			rhs := reflect.ValueOf(args[1])
			rhstype := rhs.Type()
			lhstype := ptr.ReflectTarget.Type()
			//P("rhstype = %#v, lhstype = %#v", rhstype, lhstype)
			if lhstype == rhstype {
				// have to exclude *SexpHash and *SexpReflect from this
				switch args[1].(type) {
				case *SexpHash:
					// handle below
				//case *SexpReflect:
				// handle here or below?
				default:
					//P("we have a reflection capable type match!")
					ptr.ReflectTarget.Elem().Set(rhs.Elem())
					return
				}
			}

			//P("derefSet: arg0 is %T and arg1 is %T,   ptr.Target = %#v", args[0], args[1], ptr.Target)
			//P("args[0] has ptr.ReflectTarget = '%#v'", ptr.ReflectTarget)
			switch payload := args[1].(type) {
			case *SexpInt:
				//Q("ptr = '%#v'", ptr)
				//Q("ptr.ReflectTarget = '%#v'", ptr.ReflectTarget)
				//Q("ptr.ReflectTarget.CanAddr() = '%#v'", ptr.ReflectTarget.Elem().CanAddr())
				//Q("ptr.ReflectTarget.CanSet() = '%#v'", ptr.ReflectTarget.Elem().CanSet())
				//Q("*SexpInt case: payload = '%#v'", payload)
				vo := reflect.ValueOf(payload.Val)
				vot := vo.Type()
				if !vot.AssignableTo(ptr.ReflectTarget.Elem().Type()) {
					return SexpNull, fmt.Errorf("type mismatch: value of type '%s' is not assignable to type '%v'",
						vot, ptr.ReflectTarget.Elem().Type())
				}
				ptr.ReflectTarget.Elem().Set(vo)
				return
			case *SexpStr:
				vo := reflect.ValueOf(payload.S)
				vot := vo.Type()
				//P("payload is *SexpStr")
				//tele := ptr.ReflectTarget.Elem()
				//P("ptr = %#v", ptr)
				tele := ptr.ReflectTarget
				//P("got past tele : %#v", tele)
				if !reflect.PtrTo(vot).AssignableTo(tele.Type()) {
					return SexpNull, fmt.Errorf("type mismatch: value of type '%v' is not assignable to '%v'",
						vot, ptr.PointedToType.RegisteredName) // tele.Type())
				}
				//P("payload is *SexpStr, got past type check")
				ptr.ReflectTarget.Elem().Set(vo)
				return
			case *SexpHash:
				//P("ptr.PointedToType = '%#v'", ptr.PointedToType)
				pt := payload.Type()
				tt := ptr.PointedToType
				if tt == pt && tt.RegisteredName == pt.RegisteredName {
					//P("have matching type!: %v", tt.RegisteredName)
					ptr.Target.(*SexpHash).CloneFrom(payload)
					return
				} else {
					return SexpNull, fmt.Errorf("cannot assign type '%v' to type '%v'",
						payload.Type().RegisteredName,
						ptr.PointedToType.RegisteredName)
				}

			case *SexpReflect:
				//Q("good, e2 is SexpReflect with Val='%#v'", payload.Val)

				//Q("ptr.Target = '%#v'.  ... trying SexpToGoStructs()", ptr.Target)
				iface, err := SexpToGoStructs(payload, ptr.Target, env, nil, 0, ptr.Target)
				_ = iface
				if err != nil {
					return SexpNull, err
				}
				//Q("got back iface = '%#v'", iface)
				panic("not done yet with this implementation of args[1] of type *SexpReflect")
			}
			return SexpNull, fmt.Errorf("derefSet doesn't handle assignment of type %T at present", args[1])

		default:
			return SexpNull, fmt.Errorf("unimplemented operation '%s' in DerefFunction", name)
		}
	}
}

// "." dot operator
func DotFunction(env *Zlisp, name string, args []Sexp) (Sexp, error) {
	if len(args) != 2 {
		return SexpNull, WrongNargs
	}
	P("in DotFunction(), name='%v', args[0] = '%v', args[1]= '%v'",
		name,
		args[0].SexpString(nil),
		args[1].SexpString(nil))
	return SexpNull, nil
	/*
		var ret Sexp = SexpNull
		var err error
		lenpath := len(path)

		if lenpath == 1 && setVal != nil {
			// single path element set, bind it now.
			a := path[0][1:] // strip off the dot
			asym := env.MakeSymbol(a)

			// check conflict
			//Q("asym = %#v\n", asym)
			builtin, typ := env.IsBuiltinSym(asym)
			if builtin {
				return SexpNull, fmt.Errorf("'%s' is a %s, cannot assign to it with dot-symbol", asym.name, typ)
			}

			err := env.LexicalBindSymbol(asym, *setVal)
			if err != nil {
				return SexpNull, err
			}
			return *setVal, nil
		}

		// handle multiple paths that index into hashes after the
		// the first

		key := path[0][1:] // strip off the dot
		//Q("\n in dotGetSetHelper(), looking up '%s'\n", key)
		ret, err, _ = env.LexicalLookupSymbol(env.MakeSymbol(key), false)
		if err != nil {
			//Q("\n in dotGetSetHelper(), '%s' not found\n", key)
			return SexpNull, err
		}
		return ret, err
	*/
}

func stripAnyDotPrefix(s string) string {
	if len(s) > 0 && s[0] == '.' {
		return s[1:]
	}
	return s
}

// SubstituteRHS locates any SexpSelector(s) (Selector implementers, really)
// and substitutes
// the value of x.RHS() for each x in args.
func (env *Zlisp) SubstituteRHS(args []Sexp) ([]Sexp, error) {
	for i := range args {
		obj, hasRhs := args[i].(Selector)
		if hasRhs {
			sx, err := obj.RHS(env)
			if err != nil {
				return args, err
			}
			args[i] = sx
		}
	}
	return args, nil
}

func ScriptFacingRegisterDemoStructs(env *Zlisp, name string, args []Sexp) (Sexp, error) {
	RegisterDemoStructs()
	return SexpNull, nil
}

func GetEnvFunction(name string) ZlispUserFunction {
	return func(env *Zlisp, _ string, args []Sexp) (Sexp, error) {
		narg := len(args)
		//fmt.Printf("GetEnv name='%s' called with narg = %v\n", name, narg)
		if name == "getenv" {
			if narg != 1 {
				return SexpNull, WrongNargs
			}
		} else {
			if name != "setenv" {
				panic("only getenv or setenv allowed here")
			}
			if narg != 2 {
				return SexpNull, WrongNargs
			}
		}
		nm := make([]string, narg)
		for i := 0; i < narg; i++ {
			switch x := args[i].(type) {
			case *SexpSymbol:
				nm[i] = x.name
			case *SexpStr:
				nm[i] = x.S
			default:
				return SexpNull, fmt.Errorf("symbol or string required, but saw %T/%v for i=%v arg", args[i], args[i].SexpString(nil), i)
			}
		}

		if name == "getenv" {
			return &SexpStr{S: os.Getenv(nm[0])}, nil
		}

		//fmt.Printf("calling setenv with nm[0]='%s', nm[1]='%s'\n", nm[0], nm[1])
		return SexpNull, os.Setenv(nm[0], nm[1])
	}
}

// coerce numbers to uint64
func AsUint64Function(env *Zlisp, name string, args []Sexp) (Sexp, error) {
	if len(args) != 1 {
		return SexpNull, WrongNargs
	}

	var val uint64
	switch x := args[0].(type) {
	case *SexpInt:
		val = uint64(x.Val)
	case *SexpFloat:
		val = uint64(x.Val)
	default:
		return SexpNull, fmt.Errorf("Cannot convert %s to uint64", TypeOf(args[0]).SexpString(nil))

	}
	return &SexpUint64{Val: val}, nil
}
