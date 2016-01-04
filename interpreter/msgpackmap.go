package glisp

import (
	"errors"
	"fmt"
)

func (env *Glisp) ImportMsgpackMap() {
	env.AddMacro("msgpack-map", MsgpackMapMacro)
	env.AddFunction("declare-msgpack-map", DeclareMsgpackMapFunction)
}

// (req path) avoids the need to put quotes around path you are sourcing.
func MsgpackMapMacro(env *Glisp, name string,
	args []Sexp) (Sexp, error) {

	if len(args) < 1 {
		return SexpNull, fmt.Errorf("struct-name is missing. use: " +
			"(msgpack-map struct-name)\n")
	}

	//	return MakeList([]Sexp{env.MakeSymbol("declare-msgpack-map"),
	//		SexpStr(args[0].(SexpSymbol).name)}), nil
	return MakeList([]Sexp{
		env.MakeSymbol("def"),
		args[0],
		MakeList([]Sexp{
			env.MakeSymbol("quote"),
			env.MakeSymbol("hash"),
		}),
	}), nil
}

func DeclareMsgpackMapFunction(env *Glisp, name string, args []Sexp) (Sexp, error) {
	if len(args) != 1 {
		return SexpNull, WrongNargs
	}

	switch t := args[0].(type) {
	case SexpStr:
		return SexpStr(string(t)), nil
	}
	return SexpNull, errors.New("argument must be string: the name of the new msgpack-map constructor function to create")
}
