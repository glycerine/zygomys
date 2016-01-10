package zygo

import (
	"errors"
)

type SexpCoroutine struct {
	env *Glisp
}

func (coro SexpCoroutine) SexpString() string {
	return "[coroutine]"
}

func StartCoroutineFunction(env *Glisp, name string,
	args []Sexp) (Sexp, error) {
	switch t := args[0].(type) {
	case SexpCoroutine:
		go t.env.Run()
	default:
		return SexpNull, errors.New("not a coroutine")
	}
	return SexpNull, nil
}

func CreateCoroutineMacro(env *Glisp, name string,
	args []Sexp) (Sexp, error) {
	coroenv := env.Duplicate()
	err := coroenv.LoadExpressions(args)
	if err != nil {
		return SexpNull, nil
	}
	coro := SexpCoroutine{coroenv}

	// (apply StartCoroutineFunction [coro])
	return MakeList([]Sexp{env.MakeSymbol("apply"),
		MakeUserFunction("__start", StartCoroutineFunction),
		SexpArray([]Sexp{coro})}), nil
}

func ImportCoroutines(env *Glisp) {
	env.AddMacro("go", CreateCoroutineMacro)
}
