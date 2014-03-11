package glisp

import (
	"errors"
	"bytes"
	"fmt"
)

var WrongNargs error = errors.New("wrong number of arguments")

type GlispFunction []Instruction
type GlispUserFunction func(*Glisp, string, []Sexp) (Sexp, error)

func (f GlispFunction) SexpString() string {
	return "function"
}

func (f GlispUserFunction) SexpString() string {
	return "user_function"
}

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

	return SexpPair{args[0], args[1]}, nil
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
	newenv.mainfunc = GlispFunction(gen.instructions)
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

var BuiltinFunctions = map[string]GlispUserFunction {
	"<" : CompareFunction,
	">" : CompareFunction,
	"<=": CompareFunction,
	">=": CompareFunction,
	"=" : CompareFunction,
	"not=": CompareFunction,
	"sll": BinaryIntFunction,
	"sra": BinaryIntFunction,
	"srl": BinaryIntFunction,
	"mod": BinaryIntFunction,
	"+": NumericFunction,
	"-": NumericFunction,
	"*": NumericFunction,
	"/": NumericFunction,
	"bit-and": BitwiseFunction,
	"bit-or": BitwiseFunction,
	"bit-xor": BitwiseFunction,
	"read": ReadFunction,
	"cons": ConsFunction,
	"first": FirstFunction,
	"rest": RestFunction,
	"list?": TypeQueryFunction,
	"null?": TypeQueryFunction,
	"array?": TypeQueryFunction,
	"number?": TypeQueryFunction,
	"int?": TypeQueryFunction,
	"float?": TypeQueryFunction,
	"char?": TypeQueryFunction,
	"symbol?": TypeQueryFunction,
	"println": PrintFunction,
	"print": PrintFunction,
}
