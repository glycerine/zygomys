package glisp

import (
	"bytes"
	"errors"
	"fmt"
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
	case "=":
		cond = res == 0
	case "not=":
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
		return expr[0], nil
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

func AgetFunction(env *Glisp, name string, args []Sexp) (Sexp, error) {
	if len(args) != 2 {
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

	return arr[i], nil
}

func AsetFunction(env *Glisp, name string, args []Sexp) (Sexp, error) {
	if len(args) != 3 {
		return SexpNull, WrongNargs
	}

	var arr SexpArray
	switch t := args[0].(type) {
	case SexpArray:
		arr = t
	default:
		return SexpNull, errors.New("First argument of aset must be array")
	}

	var i int
	switch t := args[1].(type) {
	case SexpInt:
		i = int(t)
	case SexpChar:
		i = int(t)
	default:
		return SexpNull, errors.New("Second argument of aset must be integer")
	}

	if i < 0 || i >= len(arr) {
		return SexpNull, errors.New("Array index out of bounds")
	}

	arr[i] = args[2]

	return arr, nil
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

func HashFunction(env *Glisp, name string, args []Sexp) (Sexp, error) {
	if len(args) < 2 || len(args) > 3 {
		return SexpNull, WrongNargs
	}

	var hash SexpHash
	switch e := args[0].(type) {
	case SexpHash:
		hash = e
	default:
		return SexpNull, errors.New("first argument of hget must be hash")
	}

	switch name {
	case "hget":
		if len(args) == 3 {
			return HashGetDefault(hash, args[1], args[2])
		}
		return HashGet(hash, args[1])
	case "hset!":
		err := HashSet(hash, args[1], args[2])
		return SexpNull, err
	case "hdel!":
		if len(args) != 2 {
			return SexpNull, WrongNargs
		}
		err := HashDelete(hash, args[1])
		return SexpNull, err
	}

	return SexpNull, nil
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

	return SexpInt(0), errors.New("argument must be string or array")
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
	if len(args) != 2 {
		return SexpNull, WrongNargs
	}

	switch t := args[0].(type) {
	case SexpArray:
		return ConcatArray(t, args[1])
	case SexpStr:
		return ConcatStr(t, args[1])
	case SexpPair:
		return ConcatList(t, args[1])
	}

	return SexpNull, errors.New("expected strings or arrays")
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
	return ParseExpression(&parser)
}

func EvalFunction(env *Glisp, name string, args []Sexp) (Sexp, error) {
	if len(args) != 1 {
		return SexpNull, WrongNargs
	}
	newenv := NewGlisp()
	gen := NewGenerator(newenv)
	err := gen.Generate(args[0])
	if err != nil {
		return SexpNull, errors.New("failed to compile expression")
	}
	newenv.mainfunc = MakeFunction("__main", 0, gen.instructions)
	newenv.pc = -1
	return newenv.Run()
}

func TypeQueryFunction(env *Glisp, name string, args []Sexp) (Sexp, error) {
	if len(args) != 1 {
		return SexpNull, WrongNargs
	}

	var result bool

	switch name {
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
	if len(args) != 1 {
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

	switch e := args[0].(type) {
	case SexpFunction:
		fun = e
	default:
		return SexpNull, errors.New("first argument must be function")
	}

	switch e := args[1].(type) {
	case SexpArray:
		return MapArray(env, fun, e)
	case SexpPair:
		return MapList(env, fun, e)
	}
	return SexpNull, errors.New("second argument must be array")
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
		return MakeHash(args)
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

var MissingFunction = SexpFunction{"__missing", true, 0, nil, nil}

func MakeFunction(name string, nargs int, fun GlispFunction) SexpFunction {
	var sfun SexpFunction
	sfun.name = name
	sfun.user = false
	sfun.nargs = nargs
	sfun.fun = fun
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
	"=":          CompareFunction,
	"not=":       CompareFunction,
	"sll":        BinaryIntFunction,
	"sra":        BinaryIntFunction,
	"srl":        BinaryIntFunction,
	"mod":        BinaryIntFunction,
	"+":          NumericFunction,
	"-":          NumericFunction,
	"*":          NumericFunction,
	"/":          NumericFunction,
	"bit-and":    BitwiseFunction,
	"bit-or":     BitwiseFunction,
	"bit-xor":    BitwiseFunction,
	"bit-not":    ComplementFunction,
	"read":       ReadFunction,
	"cons":       ConsFunction,
	"first":      FirstFunction,
	"rest":       RestFunction,
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
	"not":        NotFunction,
	"apply":      ApplyFunction,
	"map":        MapFunction,
	"make-array": MakeArrayFunction,
	"aget":       AgetFunction,
	"aset!":      AsetFunction,
	"sget":       SgetFunction,
	"hget":       HashFunction,
	"hset!":      HashFunction,
	"hdel!":      HashFunction,
	"slice":      SliceFunction,
	"len":        LenFunction,
	"append":     AppendFunction,
	"concat":     ConcatFunction,
	"array":      ConstructorFunction,
	"list":       ConstructorFunction,
	"hash":       ConstructorFunction,
	"symnum":     SymnumFunction,
}
