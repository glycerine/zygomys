package zygo

import (
	"bytes"
	"errors"
	"fmt"
	"os"
)

var WrongNargs error = errors.New("wrong number of arguments")

type GlispFunction []Instruction
type GlispUserFunction func(*Glisp, string, []Sexp) (Sexp, error)

func CompareFunction(env *Glisp, name string, args []Sexp) (Sexp, error) {
	if len(args) != 2 {
		return SexpNull, WrongNargs
	}

	res, err := Compare(args[0], args[1])
	if err != nil {
		return SexpNull, err
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
	case "not=":
		cond = res != 0
	case "!=":
		cond = res != 0
	}

	return SexpBool(cond), nil
}

func BinaryIntFunction(env *Glisp, name string, args []Sexp) (Sexp, error) {
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

func BitwiseFunction(env *Glisp, name string, args []Sexp) (Sexp, error) {
	if len(args) != 2 {
		return SexpNull, WrongNargs
	}

	var op IntegerOp
	switch name {
	case "bit-and":
		op = BitAnd
	case "bit-or":
		op = BitOr
	case "bit-xor":
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

func ComplementFunction(env *Glisp, name string, args []Sexp) (Sexp, error) {
	if len(args) != 1 {
		return SexpNull, WrongNargs
	}

	switch t := args[0].(type) {
	case SexpInt:
		return ^t, nil
	case SexpChar:
		return ^t, nil
	}

	return SexpNull, errors.New("Argument to bit-not should be integer")
}

func NumericFunction(env *Glisp, name string, args []Sexp) (Sexp, error) {
	if len(args) < 1 {
		return SexpNull, WrongNargs
	}

	var err error
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

func ConsFunction(env *Glisp, name string, args []Sexp) (Sexp, error) {
	if len(args) != 2 {
		return SexpNull, WrongNargs
	}

	return Cons(args[0], args[1]), nil
}

func FirstFunction(env *Glisp, name string, args []Sexp) (Sexp, error) {
	if len(args) != 1 {
		return SexpNull, WrongNargs
	}
	switch expr := args[0].(type) {
	case SexpPair:
		return expr.Head, nil
	case SexpArray:
		if len(expr) > 0 {
			return expr[0], nil
		}
		return SexpNull, fmt.Errorf("first called on empty array")
	}
	return SexpNull, WrongType
}

func RestFunction(env *Glisp, name string, args []Sexp) (Sexp, error) {
	if len(args) != 1 {
		return SexpNull, WrongNargs
	}

	switch expr := args[0].(type) {
	case SexpPair:
		return expr.Tail, nil
	case SexpArray:
		if len(expr) == 0 {
			return expr, nil
		}
		return expr[1:], nil
	case SexpSentinel:
		if expr == SexpNull {
			return SexpNull, nil
		}
	}

	return SexpNull, WrongType
}

func SecondFunction(env *Glisp, name string, args []Sexp) (Sexp, error) {
	if len(args) != 1 {
		return SexpNull, WrongNargs
	}
	switch expr := args[0].(type) {
	case SexpPair:
		tail := expr.Tail
		switch p := tail.(type) {
		case SexpPair:
			return p.Head, nil
		}
		return SexpNull, fmt.Errorf("list too small for second")
	case SexpArray:
		if len(expr) >= 2 {
			return expr[1], nil
		}
		return SexpNull, fmt.Errorf("array too small for second")
	}

	return SexpNull, WrongType
}

func ArrayAccessFunction(env *Glisp, name string, args []Sexp) (Sexp, error) {
	narg := len(args)
	if narg < 2 || narg > 3 {
		return SexpNull, WrongNargs
	}

	var arr SexpArray
	switch t := args[0].(type) {
	case SexpArray:
		arr = t
	default:
		return SexpNull, errors.New("First argument of aget must be array")
	}

	var i int
	switch t := args[1].(type) {
	case SexpInt:
		i = int(t)
	case SexpChar:
		i = int(t)
	default:
		// can we evaluate it?
		res, err := EvalFunction(env, "eval-aget-index", []Sexp{args[1]})
		if err != nil {
			return SexpNull, fmt.Errorf("error during eval of "+
				"array-access position argument: %s", err)
		}
		switch j := res.(type) {
		case SexpInt:
			i = int(j)
		default:
			return SexpNull, errors.New("Second argument of aget could not be evaluated to integer")
		}
	}

	switch name {
	case "hget":
		fallthrough
	case "aget":
		if i < 0 || i >= len(arr) {
			// out of bounds -- do we have a default?
			if narg == 3 {
				return args[2], nil
			}
			return SexpNull, errors.New("Array index out of bounds")
		}
		return arr[i], nil
	case "aset!":
		if len(args) != 3 {
			return SexpNull, WrongNargs
		}
		arr[i] = args[2]
	}
	return SexpNull, nil
}

func SgetFunction(env *Glisp, name string, args []Sexp) (Sexp, error) {
	if len(args) != 2 {
		return SexpNull, WrongNargs
	}

	var str SexpStr
	switch t := args[0].(type) {
	case SexpStr:
		str = t
	default:
		return SexpNull, errors.New("First argument of sget must be string")
	}

	var i int
	switch t := args[1].(type) {
	case SexpInt:
		i = int(t)
	case SexpChar:
		i = int(t)
	default:
		return SexpNull, errors.New("Second argument of sget must be integer")
	}

	return SexpChar(str[i]), nil
}

func HashAccessFunction(env *Glisp, name string, args []Sexp) (Sexp, error) {
	if len(args) < 1 || len(args) > 3 {
		return SexpNull, WrongNargs
	}

	var hash SexpHash
	switch e := args[0].(type) {
	case SexpHash:
		hash = e
	default:
		return SexpNull, errors.New("first argument of to h* function must be hash")
	}

	switch name {
	case "hget":
		if len(args) == 3 {
			return hash.HashGetDefault(env, args[1], args[2])
		}
		return hash.HashGet(env, args[1])
	case "hset!":
		if len(args) != 3 {
			return SexpNull, WrongNargs
		}
		err := hash.HashSet(args[1], args[2])
		return SexpNull, err
	case "hdel!":
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
		n := len(*(hash.KeyOrder))
		for i := 0; i < n; i++ {
			keys = append(keys, (*hash.KeyOrder)[i])
		}
		return SexpArray(keys), nil
	case "hpair":
		if len(args) != 2 {
			return SexpNull, WrongNargs
		}
		switch posreq := args[1].(type) {
		case SexpInt:
			pos := int(posreq)
			if pos < 0 || pos >= len(*hash.KeyOrder) {
				return SexpNull, fmt.Errorf("hpair position request %d out of bounds", pos)
			}
			return hash.HashPairi(pos)
		default:
			return SexpNull, fmt.Errorf("hpair position request must be an integer")
		}
	}

	return SexpNull, nil
}

func HashColonFunction(env *Glisp, name string, args []Sexp) (Sexp, error) {
	if len(args) < 2 || len(args) > 3 {
		return SexpNull, WrongNargs
	}

	var hash SexpHash
	switch e := args[1].(type) {
	case SexpHash:
		hash = e
	default:
		return SexpNull, errors.New("second argument of (:field hash) must be a hash")
	}

	if len(args) == 3 {
		return hash.HashGetDefault(env, args[0], args[2])
	}
	return hash.HashGet(env, args[0])
}

func SliceFunction(env *Glisp, name string, args []Sexp) (Sexp, error) {
	if len(args) != 3 {
		return SexpNull, WrongNargs
	}

	var start int
	var end int
	switch t := args[1].(type) {
	case SexpInt:
		start = int(t)
	case SexpChar:
		start = int(t)
	default:
		return SexpNull, errors.New("Second argument of slice must be integer")
	}

	switch t := args[2].(type) {
	case SexpInt:
		end = int(t)
	case SexpChar:
		end = int(t)
	default:
		return SexpNull, errors.New("Third argument of slice must be integer")
	}

	switch t := args[0].(type) {
	case SexpArray:
		return SexpArray(t[start:end]), nil
	case SexpStr:
		return SexpStr(t[start:end]), nil
	}

	return SexpNull, errors.New("First argument of slice must be array or string")
}

func LenFunction(env *Glisp, name string, args []Sexp) (Sexp, error) {
	if len(args) != 1 {
		return SexpNull, WrongNargs
	}

	switch t := args[0].(type) {
	case SexpSentinel:
		if t == SexpNull {
			return SexpInt(0), nil
		}
		break
	case SexpArray:
		return SexpInt(len(t)), nil
	case SexpStr:
		return SexpInt(len(t)), nil
	case SexpHash:
		return SexpInt(HashCountKeys(t)), nil
	case SexpPair:
		n, err := ListLen(t)
		return SexpInt(n), err
	}

	return SexpInt(0), errors.New("argument must be string, list, hash, or array")
}

func AppendFunction(env *Glisp, name string, args []Sexp) (Sexp, error) {
	if len(args) != 2 {
		return SexpNull, WrongNargs
	}

	switch t := args[0].(type) {
	case SexpArray:
		return SexpArray(append(t, args[1])), nil
	case SexpStr:
		return AppendStr(t, args[1])
	}

	return SexpNull, errors.New("First argument of append must be array or string")
}

func ConcatFunction(env *Glisp, name string, args []Sexp) (Sexp, error) {
	if len(args) < 1 {
		return SexpNull, WrongNargs
	}

	switch t := args[0].(type) {
	case SexpArray:
		return ConcatArray(t, args[1:])
	case SexpStr:
		return ConcatStr(t, args[1:])
	case SexpPair:
		n := len(args)
		switch {
		case n == 2:
			return ConcatList(t, args[1])
		case n == 1:
			return t, nil
		default:
			return SexpNull, WrongNargs

		}
	}

	return SexpNull, errors.New("expected strings, lists or arrays")
}

func ReadFunction(env *Glisp, name string, args []Sexp) (Sexp, error) {
	if len(args) != 1 {
		return SexpNull, WrongNargs
	}
	str := ""
	switch t := args[0].(type) {
	case SexpStr:
		str = string(t)
	default:
		return SexpNull, WrongType
	}
	env.parser.ResetAddNewInput(bytes.NewBuffer([]byte(str)))
	exp, err := env.parser.ParseExpression(0)
	return exp, err
}

func EvalFunction(env *Glisp, name string, args []Sexp) (Sexp, error) {
	if len(args) != 1 {
		return SexpNull, WrongNargs
	}
	newenv := env.Duplicate()
	err := newenv.LoadExpressions(args)
	if err != nil {
		return SexpNull, errors.New("failed to compile expression")
	}
	newenv.pc = 0
	return newenv.Run()
}

func TypeQueryFunction(env *Glisp, name string, args []Sexp) (Sexp, error) {
	if len(args) != 1 {
		return SexpNull, WrongNargs
	}

	var result bool

	switch name {
	case "type":
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
	}

	return SexpBool(result), nil
}

func PrintFunction(env *Glisp, name string, args []Sexp) (Sexp, error) {
	if len(args) < 1 {
		return SexpNull, WrongNargs
	}

	var str string

	switch expr := args[0].(type) {
	case SexpStr:
		str = string(expr)
	default:
		str = expr.SexpString()
	}

	switch name {
	case "println":
		fmt.Println(str)
	case "print":
		fmt.Print(str)
	case "printf":
		if len(args) == 1 {
			fmt.Printf(str)
		} else {
			ar := make([]interface{}, len(args)-1)
			for i := 0; i < len(ar); i++ {
				ar[i] = args[i+1]
			}
			fmt.Printf(str, ar...)
		}
	}

	return SexpNull, nil
}

func NotFunction(env *Glisp, name string, args []Sexp) (Sexp, error) {
	if len(args) != 1 {
		return SexpNull, WrongNargs
	}

	result := SexpBool(!IsTruthy(args[0]))
	return result, nil
}

func ApplyFunction(env *Glisp, name string, args []Sexp) (Sexp, error) {
	if len(args) != 2 {
		return SexpNull, WrongNargs
	}
	var fun *SexpFunction
	var funargs SexpArray

	switch e := args[0].(type) {
	case *SexpFunction:
		fun = e
	default:
		return SexpNull, errors.New("first argument must be function")
	}

	switch e := args[1].(type) {
	case SexpArray:
		funargs = e
	case SexpPair:
		var err error
		funargs, err = ListToArray(e)
		if err != nil {
			return SexpNull, err
		}
	default:
		return SexpNull, errors.New("second argument must be array or list")
	}

	return env.Apply(fun, funargs)
}

func MapFunction(env *Glisp, name string, args []Sexp) (Sexp, error) {
	if len(args) != 2 {
		return SexpNull, WrongNargs
	}
	var fun *SexpFunction

	VPrintf("\n debug Map: args = '%#v'\n", args)

	switch e := args[0].(type) {
	case *SexpFunction:
		fun = e
	default:
		return SexpNull, fmt.Errorf("first argument must be function, but we had %T / val = '%#v'", e, e)
	}

	switch e := args[1].(type) {
	case SexpArray:
		return MapArray(env, fun, e)
	case SexpPair:
		x, err := MapList(env, fun, e)
		return x, err
	default:
		return SexpNull, fmt.Errorf("second argument must be array or list; we saw %T / val = %#v", e, e)
	}
}

func MakeArrayFunction(env *Glisp, name string, args []Sexp) (Sexp, error) {
	if len(args) < 1 {
		return SexpNull, WrongNargs
	}

	var size int
	switch e := args[0].(type) {
	case SexpInt:
		size = int(e)
	default:
		return SexpNull, errors.New("first argument must be integer")
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

	return SexpArray(arr), nil
}

func ConstructorFunction(env *Glisp, name string, args []Sexp) (Sexp, error) {
	switch name {
	case "array":
		return SexpArray(args), nil
	case "list":
		return MakeList(args), nil
	case "hash":
		return MakeHash(args, "hash", env)
	case "raw":
		return MakeRaw(args)
	case "msgmap":
		switch len(args) {
		case 0:
			return MakeHash(args, "msgmap", env)
		default:
			arr, err := ListToArray(args[1])
			if err != nil {
				return SexpNull, fmt.Errorf("error converting "+
					"msgmap arguments to an array: '%v'", err)
			}
			switch nm := args[0].(type) {
			case SexpStr:
				return MakeHash(arr, string(nm), env)
			case SexpSymbol:
				return MakeHash(arr, nm.name, env)
			default:
				return MakeHash(arr, "msgmap", env)
			}
		}
	}
	return SexpNull, errors.New("invalid constructor")
}

func SymnumFunction(env *Glisp, name string, args []Sexp) (Sexp, error) {
	if len(args) != 1 {
		return SexpNull, WrongNargs
	}

	switch t := args[0].(type) {
	case SexpSymbol:
		return SexpInt(t.number), nil
	}
	return SexpNull, errors.New("argument must be symbol")
}

var MissingFunction = &SexpFunction{name: "__missing", user: true}

func (env *Glisp) MakeFunction(name string, nargs int, varargs bool,
	fun GlispFunction, orig Sexp) *SexpFunction {
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

func MakeUserFunction(name string, ufun GlispUserFunction) *SexpFunction {
	var sfun SexpFunction
	sfun.name = name
	sfun.user = true
	sfun.userfun = ufun
	return &sfun
}

var BuiltinFunctions = map[string]GlispUserFunction{
	"<":          CompareFunction,
	">":          CompareFunction,
	"<=":         CompareFunction,
	">=":         CompareFunction,
	"==":         CompareFunction,
	"not=":       CompareFunction,
	"!=":         CompareFunction,
	"sll":        BinaryIntFunction,
	"sra":        BinaryIntFunction,
	"srl":        BinaryIntFunction,
	"mod":        BinaryIntFunction,
	"+":          NumericFunction,
	"-":          NumericFunction,
	"*":          NumericFunction,
	"**":         NumericFunction,
	"/":          NumericFunction,
	"bit-and":    BitwiseFunction,
	"bit-or":     BitwiseFunction,
	"bit-xor":    BitwiseFunction,
	"bit-not":    ComplementFunction,
	"read":       ReadFunction,
	"cons":       ConsFunction,
	"first":      FirstFunction,
	"second":     SecondFunction,
	"rest":       RestFunction,
	"car":        FirstFunction,
	"cdr":        RestFunction,
	"type":       TypeQueryFunction,
	"list?":      TypeQueryFunction,
	"null?":      TypeQueryFunction,
	"array?":     TypeQueryFunction,
	"hash?":      TypeQueryFunction,
	"number?":    TypeQueryFunction,
	"int?":       TypeQueryFunction,
	"float?":     TypeQueryFunction,
	"char?":      TypeQueryFunction,
	"symbol?":    TypeQueryFunction,
	"string?":    TypeQueryFunction,
	"zero?":      TypeQueryFunction,
	"empty?":     TypeQueryFunction,
	"println":    PrintFunction,
	"print":      PrintFunction,
	"printf":     PrintFunction,
	"not":        NotFunction,
	"apply":      ApplyFunction,
	"map":        MapFunction,
	"make-array": MakeArrayFunction,
	"aget":       ArrayAccessFunction,
	"aset!":      ArrayAccessFunction,
	"sget":       SgetFunction,
	"hget":       GenericAccessFunction, // handles arrays or hashes
	"hset!":      HashAccessFunction,
	"hdel!":      HashAccessFunction,
	"keys":       HashAccessFunction,
	"hpair":      GenericHpairFunction,
	"slice":      SliceFunction,
	"len":        LenFunction,
	"append":     AppendFunction,
	"concat":     ConcatFunction,
	"array":      ConstructorFunction,
	"list":       ConstructorFunction,
	"hash":       ConstructorFunction,
	"msgmap":     ConstructorFunction,
	"raw":        ConstructorFunction,
	"raw2str":    RawToStringFunction,
	"symnum":     SymnumFunction,
	"source":     SourceFileFunction,
	//"source":    SimpleSourceFunction,
	"str2sym":   Str2SymFunction,
	"sym2str":   Sym2StrFunction,
	"gensym":    GensymFunction,
	"str":       StringifyFunction,
	"->":        ThreadMapFunction,
	"json":      JsonFunction,
	"unjson":    JsonFunction,
	"msgpack":   JsonFunction,
	"unmsgpack": JsonFunction,
	"togo":      ToGoFunction,
	"dump":      GoonDumpFunction,
	"slurpf":    SlurpfileFunction,
	"writef":    WriteToFileFunction,
	"owritef":   WriteToFileFunction,
	"system":    SystemFunction,
	"flatten":   FlattenToWordsFunction,
	"nsplit":    SplitStringOnNewlinesFunction,
	"methodls":  GoMethodListFunction,
	"_method":   CallGoMethodFunction,
	"fieldls":   GoFieldListFunction,
	"chomp":     StringUtilFunction,
	"exit":      ExitFunction,
	"stop":      StopFunction,
	"_closdump": DumpClosureEnvFunction,
	"_call":     CallZMethodOnRecordFunction,
	"gob":       GobEncodeFunction,
	"dot":       DotFunction,
	".":         DotFunction,
	"=":         AssignmentFunction,
	"joinsym":   JoinSymFunction,
	"quotelist": QuoteListFunction,
}

func ThreadMapFunction(env *Glisp, name string, args []Sexp) (Sexp, error) {
	if len(args) < 2 {
		return SexpNull, WrongNargs
	}

	h, isHash := args[0].(SexpHash)
	if !isHash {
		return SexpNull, fmt.Errorf("-> error: first argument must be a hash or defmap")
	}

	field, err := threadingHelper(env, &h, args[1:])
	if err != nil {
		return SexpNull, err
	}

	return field, nil
}

func threadingHelper(env *Glisp, hash *SexpHash, args []Sexp) (Sexp, error) {
	if len(args) == 0 {
		panic("should not recur without arguments")
	}
	field, err := hash.HashGet(env, args[0])
	if err != nil {
		return SexpNull, fmt.Errorf("-> error: field '%s' not found",
			args[0].SexpString())
	}
	if len(args) > 1 {
		h, isHash := field.(SexpHash)
		if !isHash {
			return SexpNull, fmt.Errorf("request for field '%s' was "+
				"not on a hash or defmap; instead type %T with value '%#v'",
				args[1].SexpString(), field, field)
		}
		return threadingHelper(env, &h, args[1:])
	}
	return field, nil
}

func StringifyFunction(env *Glisp, name string, args []Sexp) (Sexp, error) {
	if len(args) != 1 {
		return SexpNull, WrongNargs
	}

	return SexpStr(args[0].SexpString()), nil
}

func Sym2StrFunction(env *Glisp, name string, args []Sexp) (Sexp, error) {
	if len(args) != 1 {
		return SexpNull, WrongNargs
	}

	switch t := args[0].(type) {
	case SexpSymbol:
		r := SexpStr(t.name)
		return r, nil
	}
	return SexpNull, errors.New("argument must be symbol")
}

func Str2SymFunction(env *Glisp, name string, args []Sexp) (Sexp, error) {
	if len(args) != 1 {
		return SexpNull, WrongNargs
	}

	switch t := args[0].(type) {
	case SexpStr:
		return env.MakeSymbol(string(t)), nil
	}
	return SexpNull, errors.New("argument must be string")
}

func GensymFunction(env *Glisp, name string, args []Sexp) (Sexp, error) {
	n := len(args)
	switch {
	case n == 0:
		return env.GenSymbol("__gensym"), nil
	case n == 1:
		switch t := args[0].(type) {
		case SexpStr:
			return env.GenSymbol(string(t)), nil
		}
		return SexpNull, errors.New("argument must be string")
	default:
		return SexpNull, WrongNargs
	}
}

func ExitFunction(env *Glisp, name string, args []Sexp) (Sexp, error) {
	if len(args) != 1 {
		return SexpNull, WrongNargs
	}
	switch e := args[0].(type) {
	case SexpInt:
		os.Exit(int(e))
	}
	return SexpNull, errors.New("argument must be int (the exit code)")
}

// handles arrays or hashes
func GenericAccessFunction(env *Glisp, name string, args []Sexp) (Sexp, error) {
	if len(args) < 1 || len(args) > 3 {
		return SexpNull, WrongNargs
	}

	switch args[0].(type) {
	case SexpHash:
		return HashAccessFunction(env, name, args)
	case SexpArray:
		return ArrayAccessFunction(env, name, args)
	}
	return SexpNull, errors.New("first argument of to hget function must be hash or array")
}

var stopErr error = fmt.Errorf("stop")

func StopFunction(env *Glisp, name string, args []Sexp) (Sexp, error) {
	narg := len(args)
	if narg > 1 {
		return SexpNull, WrongNargs
	}

	if narg == 0 {
		return SexpNull, stopErr
	}

	switch s := args[0].(type) {
	case SexpStr:
		return SexpNull, fmt.Errorf(string(s))
	}
	return SexpNull, stopErr
}

var DotSexpFunc = &SexpFunction{
	name:    "dot",
	user:    true,
	nargs:   1,
	varargs: true,
	userfun: DotFunction,
}

// dot : object-oriented style calls
func DotFunction(env *Glisp, name string, args []Sexp) (Sexp, error) {
	P("\n DotFunction called! args='%s'\n", SexpArray(args).SexpString())

	narg := len(args)

	if narg == 0 {
		// a get request, just return the nested object
		return dotGetSetHelper(env, name, nil)
	}
	var fun *SexpFunction
	switch f := args[0].(type) {
	case *SexpFunction:
		fun = f
	default:
		return SexpNull, fmt.Errorf("method '%s' for dotcall "+
			"was not an SexpFunction", args[0].SexpString())
	}

	var err error

	if fun.user {
		P("\n user function (Go code)\n")
		// push our args, set up the call
		env.datastack.PushExpr(env.MakeDotSymbol(name))
		callargs := args[1:]
		ncallarg := len(callargs)
		P("callargs = %#v\n", callargs)
		for _, val := range callargs {
			env.datastack.PushExpr(val)
		}
		_, err = env.CallUserFunction(fun, fun.name, ncallarg+1)
	} else {
		P("\n sexp function, not user\n")
		fmt.Printf("\n before CallFunction() DataStack: (length %d)\n", env.datastack.Size())
		//env.datastack.PrintStack()

		// push our args, set up the call
		env.datastack.PushExpr(env.MakeDotSymbol(name))
		callargs := args[1:]
		ncallarg := len(callargs)
		P("callargs = %#v\n", callargs)
		for _, val := range callargs {
			env.datastack.PushExpr(val)
		}

		P("\n DotFunction calling env.CallFunction(fun='%s',%v)\n", fun.name, ncallarg+1)
		err = env.CallFunction(fun, ncallarg+1)

		fmt.Printf("\n after CallFunction() DataStack: (length %d)\n", env.datastack.Size())
		//env.datastack.PrintStack()

	}

	if err != nil {
		return SexpNull, err
	}

	return SexpNull, nil
}

// the assignment function, =
func AssignmentFunction(env *Glisp, name string, args []Sexp) (Sexp, error) {
	P("\n AssignmentFunction called with name ='%s'. args='%s'\n", name,
		SexpArray(args).SexpString())

	narg := len(args)
	if narg != 2 {
		return SexpNull, fmt.Errorf("assignment with '=' requires 2 args: lhs and rhs")
	}

	var sym SexpSymbol
	switch s := args[0].(type) {
	case SexpSymbol:
		sym = s
	default:
		return SexpNull, fmt.Errorf("assignment with '=' needs left-hand-side"+
			" argument to be a symbol; we got %T", s)
	}

	if !sym.isDot {
		P("assignment sees LHS symbol but is not dot, binding '%s' to '%s'\n",
			sym.name, args[1].SexpString())
		err := env.LexicalBindSymbol(sym, args[1])
		if err != nil {
			return SexpNull, err
		}
		return args[1], nil
	}

	/*
		path := DotPartsRegex.FindAllString(args[0], -1)
		P("path = '%#v' and narg=%v\n", path, narg)
		if len(path) == 0 {
			return SexpNull, fmt.Errorf("internal error: DotFunction path had zero length")
		}
	*/
	P("assignment calling dotGetSetHelper()\n")
	return dotGetSetHelper(env, sym.name, &args[1])
}

func JoinSymFunction(env *Glisp, name string, args []Sexp) (Sexp, error) {
	narg := len(args)
	if narg == 0 {
		return SexpNull, nil
	}

	j := ""

	for k := range args {
		switch a := args[k].(type) {
		case SexpPair:
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

		case SexpSymbol:
			j = j + a.name
		case SexpArray:
			s, err := joinSymHelper(a)
			if err != nil {
				return SexpNull, err
			}
			j += s
		default:
			return SexpNull, fmt.Errorf("error cannot joinsym type '%T' / val = '%s'", a, a.SexpString())
		}
	}

	return env.MakeSymbol(j), nil
}

func joinSymHelper(arr []Sexp) (string, error) {
	j := ""
	for i := 0; i < len(arr); i++ {
		switch s := arr[i].(type) {
		case SexpSymbol:
			j = j + s.name

		default:
			return "", fmt.Errorf("not a symbol: '%s'",
				arr[i].SexpString())
		}
	}
	return j, nil
}

// '(a b c) -> ('a 'b 'c)
func QuoteListFunction(env *Glisp, name string, args []Sexp) (Sexp, error) {
	narg := len(args)
	if narg != 1 {
		return SexpNull, WrongNargs
	}

	pair, ok := args[0].(SexpPair)
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

// if setVal is nil, only get and return the lookup.
func dotGetSetHelper(env *Glisp, name string, setVal *Sexp) (Sexp, error) {
	path := DotPartsRegex.FindAllString(name, -1)
	P("\n in dotGetSetHelper(), path = '%#v'\n", path)
	if len(path) == 0 {
		return SexpNull, fmt.Errorf("internal error: DotFunction" +
			" path had zero length")
	}

	var ret Sexp = SexpNull
	var err error
	lenpath := len(path)

	if lenpath == 1 && setVal != nil {
		// single path element set, bind it now.
		a := path[0][1:] // strip off the dot
		asym := env.MakeSymbol(a)
		err := env.LexicalBindSymbol(asym, *setVal)
		if err != nil {
			return SexpNull, err
		}
		return *setVal, nil
	}

	// handle multiple paths that index into hashes after the
	// the first

	key := path[0][1:] // strip off the dot
	P("\n in dotGetSetHelper(), looking up '%s'\n", key)
	ret, err, _ = env.LexicalLookupSymbol(env.MakeSymbol(key), false)
	if err != nil {
		P("\n in dotGetSetHelper(), '%s' not found\n", key)
		return SexpNull, err
	}
	if lenpath == 1 {
		// single path element get, return it.
		return ret, err
	}

	// at least .a.b if not a.b.c. etc: multiple elements,
	// where .b and after
	// will index into hashes (.a must refer to a hash);
	// proceed deeper into the hashes.
	h, isHash := ret.(SexpHash)
	if !isHash {
		return SexpNull, fmt.Errorf("not a record: cannot get "+
			"field '%s' in non-record (instead of type %T)",
			path[1][1:], ret)
	}
	// have hash: rest of path handled in hashutils.go in nestedPathGet()
	//P("\n in dotGetSetHelper(), about to call nestedPathGetSet() with"+
	//	"dotpaths = path[i+1:]='%#v\n", path[1:])
	exp, err := h.nestedPathGetSet(env, path[1:], setVal)
	if err != nil {
		return SexpNull, err
	}
	return exp, nil
}
