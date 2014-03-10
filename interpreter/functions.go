package glisp

import (
	"errors"
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

func CompareFunction(glisp *Glisp, name string, args []Sexp) (Sexp, error) {
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

func BinaryIntFunction(glisp *Glisp, name string, args []Sexp) (Sexp, error) {
	if len(args) != 2 {
		return SexpNull, WrongNargs
	}

	var op BinaryIntOp
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

	return BinaryIntDo(op, args[0], args[1])
}

func NumericFunction(glisp *Glisp, name string, args []Sexp) (Sexp, error) {
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
}
