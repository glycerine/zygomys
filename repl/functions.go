package zygo

import (
	"bytes"
	"errors"
	"fmt"
	//"github.com/shurcooL/go-goon"
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
		return expr.head, nil
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
		return expr.tail, nil
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
		tail := expr.tail
		switch p := tail.(type) {
		case SexpPair:
			return p.head, nil
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
	if len(args) < 2 {
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
		return SexpNull, errors.New("Second argument of aget must be integer")
	}

	if i < 0 || i >= len(arr) {
		return SexpNull, errors.New("Array index out of bounds")
	}

	if name == "aget" {
		return arr[i], nil
	}

	if len(args) != 3 {
		return SexpNull, WrongNargs
	}
	arr[i] = args[2]

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
			return hash.HashGetDefault(args[1], args[2])
		}
		return hash.HashGet(args[1])
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
		return hash.HashGetDefault(args[0], args[2])
	}
	return hash.HashGet(args[0])
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
	case SexpArray:
		return SexpInt(len(t)), nil
	case SexpStr:
		return SexpInt(len(t)), nil
	case SexpHash:
		return SexpInt(HashCountKeys(t)), nil
	}

	return SexpInt(0), errors.New("argument must be string, hash, or array")
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
	lexer := NewLexerFromStream(bytes.NewBuffer([]byte(str)))
	parser := Parser{lexer, env}
	exp, err := ParseExpression(&parser, 0)
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
	var fun SexpFunction
	var funargs SexpArray

	switch e := args[0].(type) {
	case SexpFunction:
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
	var fun SexpFunction

	VPrintf("\n debug Map: args = '%#v'\n", args)

	switch e := args[0].(type) {
	case SexpFunction:
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
		return MakeHash(args, "hash")
	case "raw":
		return MakeRaw(args)
	case "msgmap":
		switch len(args) {
		case 0:
			return MakeHash(args, "msgmap")
		default:
			arr, err := ListToArray(args[1])
			if err != nil {
				return SexpNull, fmt.Errorf("error converting "+
					"msgmap arguments to an array: '%v'", err)
			}
			switch nm := args[0].(type) {
			case SexpStr:
				return MakeHash(arr, string(nm))
			case SexpSymbol:
				return MakeHash(arr, nm.name)
			default:
				return MakeHash(arr, "msgmap")
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

func SourceFileFunction(env *Glisp, name string, args []Sexp) (Sexp, error) {
	if len(args) < 1 {
		return SexpNull, WrongNargs
	}

	var sourceItem func(item Sexp) error

	sourceItem = func(item Sexp) error {
		switch t := item.(type) {
		case SexpArray:
			for _, v := range t {
				if err := sourceItem(v); err != nil {
					return err
				}
			}
		case SexpPair:
			expr := item
			for expr != SexpNull {
				list := expr.(SexpPair)
				if err := sourceItem(list.head); err != nil {
					return err
				}
				expr = list.tail
			}
		case SexpStr:
			var f *os.File
			var err error

			if f, err = os.Open(string(t)); err != nil {
				return err
			}
			defer f.Close()
			if err = env.SourceFile(f); err != nil {
				return err
			}

		default:
			return fmt.Errorf("%v: Expected `string`, `list`, `array` given type %T val %v", name, item, item)
		}

		return nil
	}

	for _, v := range args {
		if err := sourceItem(v); err != nil {
			return SexpNull, err
		}
	}

	return SexpNull, nil
}

var MissingFunction = SexpFunction{name: "__missing", user: true}

func MakeFunction(name string, nargs int, varargs bool,
	fun GlispFunction, orig Sexp) SexpFunction {
	var sfun SexpFunction
	sfun.name = name
	sfun.user = false
	sfun.nargs = nargs
	sfun.varargs = varargs
	sfun.fun = fun
	sfun.orig = orig
	return sfun
}

func MakeUserFunction(name string, ufun GlispUserFunction) SexpFunction {
	var sfun SexpFunction
	sfun.name = name
	sfun.user = true
	sfun.userfun = ufun
	return sfun
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
	"hget":       HashAccessFunction,
	"hset!":      HashAccessFunction,
	"hdel!":      HashAccessFunction,
	"keys":       HashAccessFunction,
	"hpair":      HashAccessFunction,
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
	"str2sym":    Str2SymFunction,
	"sym2str":    Sym2StrFunction,
	"gensym":     GensymFunction,
	"str":        StringifyFunction,
	"->":         ThreadMapFunction,
	"json":       JsonFunction,
	"unjson":     JsonFunction,
	"msgpack":    JsonFunction,
	"unmsgpack":  JsonFunction,
	"togo":       ToGoFunction,
	"dump":       GoonDumpFunction,
	"slurpf":     SlurpfileFunction,
	"writef":     WriteToFileFunction,
	"owritef":    WriteToFileFunction,
	"system":     SystemFunction,
	"flatten":    FlattenToWordsFunction,
	"nsplit":     SplitStringOnNewlinesFunction,
}

func ThreadMapFunction(env *Glisp, name string, args []Sexp) (Sexp, error) {
	if len(args) < 2 {
		return SexpNull, WrongNargs
	}

	h, isHash := args[0].(SexpHash)
	if !isHash {
		return SexpNull, fmt.Errorf("-> error: first argument must be a hash or defmap")
	}

	field, err := threadingHelper(&h, args[1:])
	if err != nil {
		return SexpNull, err
	}

	return field, nil
}

func threadingHelper(hash *SexpHash, args []Sexp) (Sexp, error) {
	if len(args) == 0 {
		panic("should not recur without arguments")
	}
	field, err := hash.HashGet(args[0])
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
		return threadingHelper(&h, args[1:])
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
